package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	appContext "github.com/cloakd/common/context"
	serviceContext "github.com/cloakd/common/services"
	log "github.com/sirupsen/logrus"
)

type GeolocationResponse struct {
	IP          string  `json:"ip"`
	CountryName string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	RegionName  string  `json:"region_name"`
	CityName    string  `json:"city_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ZipCode     string  `json:"zip_code"`
	TimeZone    string  `json:"time_zone"`
	ISP         string  `json:"isp"`
}

type GeolocationService struct {
	serviceContext.DefaultService
	httpClient  *http.Client
	apiURL      string
	redisSvc    *RedisService
	cacheExpiry time.Duration
}

const GEOLOCATION_SVC = "geolocation_svc"

func (svc GeolocationService) Id() string {
	return GEOLOCATION_SVC
}

func (svc *GeolocationService) Configure(ctx *appContext.Context) error {
	svc.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	svc.apiURL = "http://ip-api.com/json"
	svc.cacheExpiry = 24 * time.Hour // Cache for 24 hours
	return svc.DefaultService.Configure(ctx)
}

func (svc *GeolocationService) Start() error {
	svc.redisSvc = svc.Service(REDIS_SVC).(*RedisService)
	return nil
}

func (svc *GeolocationService) GetLocationByIP(ip string) (string, error) {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return "Local", nil
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("geolocation:simple:%s", ip)

	// Try to get from cache first
	if svc.redisSvc != nil {
		cachedLocation, err := svc.redisSvc.Get(ctx, cacheKey)
		if err == nil && cachedLocation != "" {
			log.WithField("ip", ip).Debug("Geolocation cache hit")
			return cachedLocation, nil
		}
	}

	// Cache miss, fetch from API
	url := fmt.Sprintf("%s/%s?fields=status,country,regionName,city", svc.apiURL, ip)

	resp, err := svc.httpClient.Get(url)
	if err != nil {
		log.WithError(err).WithField("ip", ip).Error("Failed to get geolocation")
		return "Unknown", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status", resp.StatusCode).WithField("ip", ip).Error("Geolocation API returned non-200 status")
		return "Unknown", nil
	}

	var result struct {
		Status     string `json:"status"`
		Country    string `json:"country"`
		RegionName string `json:"regionName"`
		City       string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.WithError(err).WithField("ip", ip).Error("Failed to decode geolocation response")
		return "Unknown", nil
	}

	if result.Status != "success" {
		log.WithField("status", result.Status).WithField("ip", ip).Warn("Geolocation lookup failed")
		return "Unknown", nil
	}

	location := ""
	if result.City != "" {
		location = result.City
	}
	if result.RegionName != "" {
		if location != "" {
			location += ", "
		}
		location += result.RegionName
	}
	if result.Country != "" {
		if location != "" {
			location += ", "
		}
		location += result.Country
	}

	if location == "" {
		location = "Unknown"
	}

	// Cache the result
	if svc.redisSvc != nil {
		if err := svc.redisSvc.Set(ctx, cacheKey, location, svc.cacheExpiry); err != nil {
			log.WithError(err).WithField("ip", ip).Warn("Failed to cache geolocation result")
		} else {
			log.WithField("ip", ip).Debug("Geolocation result cached")
		}
	}

	return location, nil
}

func (svc *GeolocationService) GetDetailedLocationByIP(ip string) (*GeolocationResponse, error) {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return &GeolocationResponse{
			IP:          ip,
			CountryName: "Local",
			CityName:    "Local",
		}, nil
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("geolocation:detailed:%s", ip)

	// Try to get from cache first
	if svc.redisSvc != nil {
		var cachedResponse GeolocationResponse
		err := svc.redisSvc.GetJSON(ctx, cacheKey, &cachedResponse)
		if err == nil && cachedResponse.IP != "" {
			log.WithField("ip", ip).Debug("Detailed geolocation cache hit")
			return &cachedResponse, nil
		}
	}

	// Cache miss, fetch from API
	url := fmt.Sprintf("%s/%s", svc.apiURL, ip)

	resp, err := svc.httpClient.Get(url)
	if err != nil {
		log.WithError(err).WithField("ip", ip).Error("Failed to get detailed geolocation")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geolocation API returned status %d", resp.StatusCode)
	}

	var result struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Query       string  `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.WithError(err).WithField("ip", ip).Error("Failed to decode detailed geolocation response")
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("geolocation lookup failed: %s", result.Status)
	}

	geoResponse := &GeolocationResponse{
		IP:          result.Query,
		CountryName: result.Country,
		CountryCode: result.CountryCode,
		RegionName:  result.RegionName,
		CityName:    result.City,
		Latitude:    result.Lat,
		Longitude:   result.Lon,
		ZipCode:     result.Zip,
		TimeZone:    result.Timezone,
		ISP:         result.ISP,
	}

	// Cache the result
	if svc.redisSvc != nil {
		if err := svc.redisSvc.Set(ctx, cacheKey, geoResponse, svc.cacheExpiry); err != nil {
			log.WithError(err).WithField("ip", ip).Warn("Failed to cache detailed geolocation result")
		} else {
			log.WithField("ip", ip).Debug("Detailed geolocation result cached")
		}
	}

	return geoResponse, nil
}

func (svc *GeolocationService) ClearCache(ip string) error {
	if svc.redisSvc == nil {
		return fmt.Errorf("redis service not available")
	}

	ctx := context.Background()
	simpleKey := fmt.Sprintf("geolocation:simple:%s", ip)
	detailedKey := fmt.Sprintf("geolocation:detailed:%s", ip)

	return svc.redisSvc.Delete(ctx, simpleKey, detailedKey)
}

func (svc *GeolocationService) ClearAllCache() error {
	if svc.redisSvc == nil {
		return fmt.Errorf("redis service not available")
	}

	ctx := context.Background()
	keys, err := svc.redisSvc.Keys(ctx, "geolocation:*")
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return svc.redisSvc.Delete(ctx, keys...)
	}

	return nil
}

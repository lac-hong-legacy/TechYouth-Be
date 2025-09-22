package seeders

import (
	"log"
	"time"

	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"gorm.io/gorm"
)

// TimelineSeeder handles seeding historical timelines
type TimelineSeeder struct {
	db *gorm.DB
}

// NewTimelineSeeder creates a new timeline seeder
func NewTimelineSeeder(db *gorm.DB) *TimelineSeeder {
	return &TimelineSeeder{db: db}
}

// SeedTimelines seeds the database with Vietnamese historical periods
func (s *TimelineSeeder) SeedTimelines() error {
	timelines := s.getHistoricalTimelines()

	for _, timeline := range timelines {
		// Check if timeline already exists
		var existingTimeline model.Timeline
		if err := s.db.Where("id = ?", timeline.ID).First(&existingTimeline).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Timeline doesn't exist, create it
				if err := s.db.Create(&timeline).Error; err != nil {
					log.Printf("Error creating timeline %s: %v", timeline.Era, err)
					return err
				}
				log.Printf("Created timeline: %s", timeline.Era)
			} else {
				log.Printf("Error checking timeline %s: %v", timeline.Era, err)
				return err
			}
		} else {
			log.Printf("Timeline %s already exists, skipping", timeline.Era)
		}
	}

	log.Println("Timeline seeding completed successfully")
	return nil
}

// getHistoricalTimelines returns the Vietnamese historical periods
func (s *TimelineSeeder) getHistoricalTimelines() []model.Timeline {
	now := time.Now()

	timelines := []model.Timeline{
		{
			ID:          "timeline_van_lang",
			Era:         "Văn Lang",
			StartYear:   -2879,
			EndYear:     intPtr(-258),
			Description: "The legendary first Vietnamese kingdom founded by Hùng Vương. Period of Lạc Việt civilization and bronze age culture.",
			KeyEvents: jsonArray([]string{
				"Foundation of Văn Lang kingdom",
				"Establishment of Hùng dynasty",
				"Development of bronze age culture",
				"Rice cultivation advancement",
			}),
			CharacterIds: jsonArray([]string{"char_hung_vuong"}),
			ImageURL:     "/assets/timeline/van_lang.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_au_lac",
			Era:         "Âu Lạc",
			StartYear:   -257,
			EndYear:     intPtr(-207),
			Description: "Kingdom established by An Dương Vương after conquering Văn Lang. Known for the construction of Cổ Loa citadel.",
			KeyEvents: jsonArray([]string{
				"An Dương Vương conquers Văn Lang",
				"Establishment of Âu Lạc kingdom",
				"Construction of Cổ Loa citadel",
				"Introduction of crossbow technology",
			}),
			CharacterIds: jsonArray([]string{}),
			ImageURL:     "/assets/timeline/au_lac.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_bac_thuoc",
			Era:         "Bắc thuộc",
			StartYear:   -111,
			EndYear:     intPtr(938),
			Description: "Period of Chinese domination lasting over 1000 years, interrupted by several major rebellions including the Trưng Sisters and Bà Triệu.",
			KeyEvents: jsonArray([]string{
				"Chinese Han conquest (-111)",
				"Trưng Sisters rebellion (40-43 AD)",
				"Bà Triệu rebellion (248 AD)",
				"Various resistance movements",
				"Cultural and technological exchanges",
			}),
			CharacterIds: jsonArray([]string{"char_hai_ba_trung", "char_ba_trieu"}),
			ImageURL:     "/assets/timeline/bac_thuoc.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ngo",
			Era:         "Ngô",
			StartYear:   939,
			EndYear:     intPtr(965),
			Description: "First independent Vietnamese dynasty after Chinese rule, founded by Ngô Quyền following his victory at Bạch Đằng River.",
			KeyEvents: jsonArray([]string{
				"Battle of Bạch Đằng River (938)",
				"End of Chinese domination",
				"Establishment of Ngô dynasty",
				"Capital at Cổ Loa",
			}),
			CharacterIds: jsonArray([]string{"char_ngo_quyen"}),
			ImageURL:     "/assets/timeline/ngo.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_dinh_le",
			Era:         "Đinh - Tiền Lê",
			StartYear:   968,
			EndYear:     intPtr(1009),
			Description: "Period of the Đinh and Early Lê dynasties, establishing the foundation of independent Vietnamese state structure.",
			KeyEvents: jsonArray([]string{
				"Đinh Bộ Lĩnh unifies Vietnam",
				"Establishment of Đại Cồ Việt",
				"Capital moved to Hoa Lư",
				"Early Lê dynasty",
			}),
			CharacterIds: jsonArray([]string{"char_dinh_bo_linh"}),
			ImageURL:     "/assets/timeline/dinh_le.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ly",
			Era:         "Lý",
			StartYear:   1009,
			EndYear:     intPtr(1225),
			Description: "Golden age of Vietnamese culture and Buddhism. Capital moved to Thăng Long (Hanoi). Period of territorial expansion and cultural development.",
			KeyEvents: jsonArray([]string{
				"Lý Thái Tổ founds dynasty (1009)",
				"Capital moved to Thăng Long (1010)",
				"Temple of Literature established",
				"Victory over Song China",
				"Buddhist cultural flourishing",
			}),
			CharacterIds: jsonArray([]string{"char_ly_thai_to", "char_ly_thuong_kiet"}),
			ImageURL:     "/assets/timeline/ly.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_tran",
			Era:         "Trần",
			StartYear:   1225,
			EndYear:     intPtr(1400),
			Description: "Dynasty famous for successfully repelling three Mongol invasions under the leadership of Trần Hưng Đạo. Period of military innovation and cultural development.",
			KeyEvents: jsonArray([]string{
				"Trần dynasty established (1225)",
				"First Mongol invasion repelled (1258)",
				"Second Mongol invasion defeated (1285)",
				"Third Mongol invasion crushed (1287-1288)",
				"Development of guerrilla warfare tactics",
			}),
			CharacterIds: jsonArray([]string{"char_tran_hung_dao"}),
			ImageURL:     "/assets/timeline/tran.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ho",
			Era:         "Hồ",
			StartYear:   1400,
			EndYear:     intPtr(1407),
			Description: "Short-lived dynasty known for progressive reforms but ended with Chinese Ming occupation due to internal conflicts.",
			KeyEvents: jsonArray([]string{
				"Hồ Quý Ly seizes power (1400)",
				"Land and monetary reforms",
				"Internal rebellions",
				"Chinese Ming invasion (1407)",
			}),
			CharacterIds: jsonArray([]string{"char_ho_quy_ly"}),
			ImageURL:     "/assets/timeline/ho.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ming_occupation",
			Era:         "Minh chiếm đóng",
			StartYear:   1407,
			EndYear:     intPtr(1428),
			Description: "Period of Chinese Ming occupation. Vietnamese resistance led by Lê Lợi culminated in independence and the founding of the Lê dynasty.",
			KeyEvents: jsonArray([]string{
				"Chinese Ming occupation begins (1407)",
				"Cultural suppression policies",
				"Lê Lợi begins resistance (1418)",
				"Lam Sơn uprising",
				"Liberation of Vietnam (1428)",
			}),
			CharacterIds: jsonArray([]string{"char_le_loi", "char_nguyen_trai"}),
			ImageURL:     "/assets/timeline/ming_occupation.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_later_le",
			Era:         "Hậu Lê",
			StartYear:   1428,
			EndYear:     intPtr(1788),
			Description: "Longest dynasty in Vietnamese history. Golden age under Lê Thánh Tông with territorial expansion and legal codification.",
			KeyEvents: jsonArray([]string{
				"Lê dynasty established (1428)",
				"Hồng Đức golden age (1470-1497)",
				"Conquest of Champa territories",
				"Hồng Đức legal code",
				"Division into Trịnh-Nguyễn period",
			}),
			CharacterIds: jsonArray([]string{"char_le_thanh_tong"}),
			ImageURL:     "/assets/timeline/later_le.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_mac",
			Era:         "Mạc",
			StartYear:   1527,
			EndYear:     intPtr(1677),
			Description: "Dynasty that controlled northern Vietnam during the divided period, competing with the restored Lê dynasty.",
			KeyEvents: jsonArray([]string{
				"Mạc Đăng Dung seizes power (1527)",
				"Control of northern territories",
				"Conflict with restored Lê dynasty",
				"Gradual territorial losses",
			}),
			CharacterIds: jsonArray([]string{"char_mac_dang_dung"}),
			ImageURL:     "/assets/timeline/mac.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_tay_son",
			Era:         "Tây Sơn",
			StartYear:   1778,
			EndYear:     intPtr(1802),
			Description: "Revolutionary period led by the Nguyễn brothers. Quang Trung defeated Chinese Qing invasion and implemented progressive reforms.",
			KeyEvents: jsonArray([]string{
				"Tây Sơn uprising begins (1778)",
				"Defeat of Trịnh and Nguyễn lords",
				"Quang Trung becomes emperor (1788)",
				"Victory over Qing invasion (1789)",
				"Social and educational reforms",
			}),
			CharacterIds: jsonArray([]string{"char_nguyen_hue"}),
			ImageURL:     "/assets/timeline/tay_son.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_nguyen",
			Era:         "Nguyễn",
			StartYear:   1802,
			EndYear:     intPtr(1945),
			Description: "Final Vietnamese imperial dynasty. Period of unification, modernization attempts, and eventual French colonization.",
			KeyEvents: jsonArray([]string{
				"Gia Long unifies Vietnam (1802)",
				"Capital established at Huế",
				"French colonial period begins (1858)",
				"Can Vuong resistance movements",
				"End of imperial rule (1945)",
			}),
			CharacterIds: jsonArray([]string{"char_nguyen_anh", "char_nguyen_du"}),
			ImageURL:     "/assets/timeline/nguyen.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_modern",
			Era:         "Cận đại",
			StartYear:   1858,
			EndYear:     intPtr(1975),
			Description: "Period of French colonization, independence movements, and wars of liberation culminating in reunification.",
			KeyEvents: jsonArray([]string{
				"French conquest begins (1858)",
				"Nationalist movements emerge",
				"World War II and Japanese occupation",
				"Declaration of Independence (1945)",
				"First Indochina War (1946-1954)",
				"Vietnam War (1955-1975)",
				"Reunification of Vietnam (1975)",
			}),
			CharacterIds: jsonArray([]string{
				"char_phan_boi_chau",
				"char_phan_chu_trinh",
				"char_vo_thi_sau",
				"char_ho_chi_minh",
			}),
			ImageURL:  "/assets/timeline/modern.jpg",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	return timelines
}

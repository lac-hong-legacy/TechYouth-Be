package seeders

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"gorm.io/gorm"
)

// CharacterSeeder handles seeding historical characters
type CharacterSeeder struct {
	db *gorm.DB
}

// NewCharacterSeeder creates a new character seeder
func NewCharacterSeeder(db *gorm.DB) *CharacterSeeder {
	return &CharacterSeeder{db: db}
}

// SeedCharacters seeds the database with Vietnamese historical characters
func (s *CharacterSeeder) SeedCharacters() error {
	characters := s.getHistoricalCharacters()

	for _, character := range characters {
		// Check if character already exists
		var existingChar model.Character
		if err := s.db.Where("id = ?", character.ID).First(&existingChar).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Character doesn't exist, create it
				if err := s.db.Create(&character).Error; err != nil {
					log.Printf("Error creating character %s: %v", character.Name, err)
					return err
				}
				log.Printf("Created character: %s", character.Name)
			} else {
				log.Printf("Error checking character %s: %v", character.Name, err)
				return err
			}
		} else {
			log.Printf("Character %s already exists, skipping", character.Name)
		}
	}

	log.Println("Character seeding completed successfully")
	return nil
}

// getHistoricalCharacters returns the list of 20 Vietnamese historical characters
func (s *CharacterSeeder) getHistoricalCharacters() []model.Character {
	now := time.Now()

	characters := []model.Character{
		{
			ID:          "char_hung_vuong",
			Name:        "Hùng Vương",
			Dynasty:     "Văn Lang",
			Rarity:      "Legendary",
			BirthYear:   intPtr(-2879),
			DeathYear:   intPtr(-258),
			Description: "Legendary founder of the first Vietnamese state, Văn Lang. Known as the first king of Vietnam, he established the foundation of Vietnamese civilization along the Red River delta.",
			FamousQuote: "Con rồng cháu tiên",
			Achievements: jsonArray([]string{
				"Founded the Văn Lang kingdom",
				"Established the Hùng dynasty",
				"Created the foundation of Vietnamese culture",
				"Unified the Lạc Việt tribes",
			}),
			ImageURL:   "/assets/characters/hung_vuong.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ly_thai_to",
			Name:        "Lý Thái Tổ",
			Dynasty:     "Lý",
			Rarity:      "Legendary",
			BirthYear:   intPtr(974),
			DeathYear:   intPtr(1028),
			Description: "Founder of the Lý dynasty and builder of Thăng Long (modern Hanoi). He moved the capital from Hoa Lư to Thăng Long, establishing a golden age of Vietnamese culture and Buddhism.",
			FamousQuote: "Thăng Long hữu đức khí",
			Achievements: jsonArray([]string{
				"Founded the Lý dynasty in 1009",
				"Established Thăng Long as the capital",
				"Promoted Buddhism and education",
				"Unified and strengthened Vietnam",
			}),
			ImageURL:   "/assets/characters/ly_thai_to.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_tran_hung_dao",
			Name:        "Trần Hưng Đạo",
			Dynasty:     "Trần",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1228),
			DeathYear:   intPtr(1300),
			Description: "Greatest military strategist in Vietnamese history. Successfully defended Vietnam against three Mongol invasions, using guerrilla tactics and superior knowledge of local terrain.",
			FamousQuote: "Thà chết vì tổ quốc chứ không sống làm nô lệ",
			Achievements: jsonArray([]string{
				"Defeated three Mongol invasions (1258, 1285, 1287-1288)",
				"Developed innovative guerrilla warfare tactics",
				"Protected Vietnamese independence",
				"Wrote military treatises on strategy",
			}),
			ImageURL:   "/assets/characters/tran_hung_dao.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_le_loi",
			Name:        "Lê Lợi",
			Dynasty:     "Lê",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1385),
			DeathYear:   intPtr(1433),
			Description: "Hero who liberated Vietnam from Chinese Ming occupation. Founded the Lê dynasty after a ten-year resistance war, becoming Emperor Lê Thái Tổ.",
			FamousQuote: "Nam quốc sơn hà Nam đế cư",
			Achievements: jsonArray([]string{
				"Led successful rebellion against Ming China (1418-1428)",
				"Founded the Later Lê dynasty",
				"Liberated Vietnam from foreign rule",
				"Established the Proclamation of Victory over the Wu",
			}),
			ImageURL:   "/assets/characters/le_loi.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_trai",
			Name:        "Nguyễn Trãi",
			Dynasty:     "Lê",
			Rarity:      "Rare",
			BirthYear:   intPtr(1380),
			DeathYear:   intPtr(1442),
			Description: "Brilliant strategist, poet, and diplomat who served Lê Lợi. Known for his literary works and the famous 'Bình Ngô Đại Cáo' proclamation.",
			FamousQuote: "Dĩ nhân thắng bạo, dĩ nghĩa thắng phi nghĩa",
			Achievements: jsonArray([]string{
				"Authored Bình Ngô Đại Cáo (Great Proclamation of Victory)",
				"Chief advisor to Lê Lợi during independence war",
				"Renowned poet and literary figure",
				"Pioneered Vietnamese diplomatic writing",
			}),
			ImageURL:   "/assets/characters/nguyen_trai.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_hue",
			Name:        "Nguyễn Huệ (Quang Trung)",
			Dynasty:     "Tây Sơn",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1753),
			DeathYear:   intPtr(1792),
			Description: "Emperor Quang Trung, leader of the Tây Sơn rebellion. Defeated the Qing invasion and initiated important reforms including promoting Vietnamese language and culture.",
			FamousQuote: "Bắc định Thanh quân, nam bình Chúa Nguyễn",
			Achievements: jsonArray([]string{
				"Led the Tây Sơn uprising",
				"Defeated Qing army at Đống Đa (1789)",
				"Promoted Vietnamese language in education",
				"Implemented progressive social reforms",
			}),
			ImageURL:   "/assets/characters/nguyen_hue.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_hai_ba_trung",
			Name:        "Hai Bà Trưng",
			Dynasty:     "Trưng",
			Rarity:      "Legendary",
			BirthYear:   intPtr(12),
			DeathYear:   intPtr(43),
			Description: "The Trưng Sisters - Trưng Trắc and Trưng Nhị - led the first major rebellion against Chinese domination. They established an independent kingdom for three years.",
			FamousQuote: "Trước để trừ giặc, sau để cảnh báo hậu thế",
			Achievements: jsonArray([]string{
				"Led successful rebellion against Chinese Han dynasty",
				"Established independent Vietnamese kingdom (40-43 AD)",
				"Symbol of Vietnamese women's strength",
				"First recorded Vietnamese rulers after Chinese occupation",
			}),
			ImageURL:   "/assets/characters/hai_ba_trung.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ly_thuong_kiet",
			Name:        "Lý Thường Kiệt",
			Dynasty:     "Lý",
			Rarity:      "Rare",
			BirthYear:   intPtr(1019),
			DeathYear:   intPtr(1105),
			Description: "Great general of the Lý dynasty who successfully defended Vietnam against Song China. Famous for his military innovations and the poem 'Nam quốc sơn hà'.",
			FamousQuote: "Nam quốc sơn hà Nam đế cư, Tiệt nhiên định phận tại thiên thư",
			Achievements: jsonArray([]string{
				"Defeated Song Chinese invasion",
				"Authored the famous patriotic poem 'Nam quốc sơn hà'",
				"Served as regent during Lý Nhân Tông's minority",
				"Strengthened Vietnamese borders",
			}),
			ImageURL:   "/assets/characters/ly_thuong_kiet.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ngo_quyen",
			Name:        "Ngô Quyền",
			Dynasty:     "Ngô",
			Rarity:      "Rare",
			BirthYear:   intPtr(897),
			DeathYear:   intPtr(944),
			Description: "Founder of the Ngô dynasty who ended nearly 1000 years of Chinese domination. Won the decisive Battle of Bạch Đằng River against the Southern Han fleet.",
			FamousQuote: "Đại Việt độc lập",
			Achievements: jsonArray([]string{
				"Won Battle of Bạch Đằng River (938)",
				"Ended Chinese domination after 1000 years",
				"Founded the Ngô dynasty",
				"Established Vietnamese independence",
			}),
			ImageURL:   "/assets/characters/ngo_quyen.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_dinh_bo_linh",
			Name:        "Đinh Bộ Lĩnh",
			Dynasty:     "Đinh",
			Rarity:      "Rare",
			BirthYear:   intPtr(924),
			DeathYear:   intPtr(979),
			Description: "First emperor of unified Vietnam, known as Đinh Tiên Hoàng. Established the Đại Cồ Việt kingdom and brought stability after years of chaos.",
			FamousQuote: "Đại Cồ Việt Hoàng Đế",
			Achievements: jsonArray([]string{
				"Unified Vietnam under Đại Cồ Việt kingdom",
				"First emperor of independent Vietnam",
				"Established capital at Hoa Lư",
				"Created organized government structure",
			}),
			ImageURL:   "/assets/characters/dinh_bo_linh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_le_thanh_tong",
			Name:        "Lê Thánh Tông",
			Dynasty:     "Lê",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1442),
			DeathYear:   intPtr(1497),
			Description: "Greatest emperor of the Lê dynasty, known for territorial expansion and legal reforms. Created the Hồng Đức legal code and promoted literature and arts.",
			FamousQuote: "Minh đức tân dân",
			Achievements: jsonArray([]string{
				"Created the comprehensive Hồng Đức legal code",
				"Expanded territory to Champa kingdom",
				"Promoted Confucian education and civil service",
				"Golden age of Vietnamese literature and arts",
			}),
			ImageURL:   "/assets/characters/le_thanh_tong.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ho_quy_ly",
			Name:        "Hồ Quý Ly",
			Dynasty:     "Hồ",
			Rarity:      "Common",
			BirthYear:   intPtr(1336),
			DeathYear:   intPtr(1407),
			Description: "Controversial reformer who founded the short-lived Hồ dynasty. Known for progressive reforms but also for the events that led to Ming Chinese invasion.",
			FamousQuote: "Cải cách duy tân",
			Achievements: jsonArray([]string{
				"Implemented land redistribution reforms",
				"Promoted paper currency system",
				"Advanced agricultural techniques",
				"Founded Hồ dynasty (1400-1407)",
			}),
			ImageURL:   "/assets/characters/ho_quy_ly.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_mac_dang_dung",
			Name:        "Mạc Đăng Dung",
			Dynasty:     "Mạc",
			Rarity:      "Common",
			BirthYear:   intPtr(1483),
			DeathYear:   intPtr(1541),
			Description: "Founder of the Mạc dynasty who seized power from the Lê dynasty. Known for his military skills but his legitimacy was often disputed.",
			FamousQuote: "Thiên mệnh tại ta",
			Achievements: jsonArray([]string{
				"Founded the Mạc dynasty",
				"Skilled military commander",
				"Controlled northern Vietnam",
				"Established diplomatic relations with China",
			}),
			ImageURL:   "/assets/characters/mac_dang_dung.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_anh",
			Name:        "Nguyễn Ánh (Gia Long)",
			Dynasty:     "Nguyễn",
			Rarity:      "Rare",
			BirthYear:   intPtr(1762),
			DeathYear:   intPtr(1820),
			Description: "Founded the Nguyễn dynasty and unified Vietnam under Emperor Gia Long. Established Huế as the capital and created the modern borders of Vietnam.",
			FamousQuote: "Thống nhất giang sơn",
			Achievements: jsonArray([]string{
				"Unified Vietnam from north to south",
				"Founded the Nguyễn dynasty (1802-1945)",
				"Established Huế as imperial capital",
				"Created modern Vietnamese territorial unity",
			}),
			ImageURL:   "/assets/characters/nguyen_anh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_phan_boi_chau",
			Name:        "Phan Bội Châu",
			Dynasty:     "Cận đại",
			Rarity:      "Rare",
			BirthYear:   intPtr(1867),
			DeathYear:   intPtr(1940),
			Description: "Pioneering nationalist and independence activist. Led early resistance movements against French colonial rule and promoted modern education and reform.",
			FamousQuote: "Việt Nam vong quốc sử",
			Achievements: jsonArray([]string{
				"Founded Việt Nam Duy Tân Hội",
				"Led Đông Du movement to Japan",
				"Wrote influential nationalist literature",
				"Pioneer of Vietnamese independence movement",
			}),
			ImageURL:   "/assets/characters/phan_boi_chau.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_phan_chu_trinh",
			Name:        "Phan Châu Trinh",
			Dynasty:     "Cận đại",
			Rarity:      "Common",
			BirthYear:   intPtr(1872),
			DeathYear:   intPtr(1926),
			Description: "Reformist scholar who advocated for modernization and education. Promoted peaceful resistance and cultural reform during French colonial period.",
			FamousQuote: "Dân trí là gốc của mọi việc",
			Achievements: jsonArray([]string{
				"Advocated for educational reform",
				"Promoted peaceful resistance methods",
				"Founded modern Vietnamese journalism",
				"Influenced intellectual awakening movement",
			}),
			ImageURL:   "/assets/characters/phan_chu_trinh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_du",
			Name:        "Nguyễn Du",
			Dynasty:     "Nguyễn",
			Rarity:      "Rare",
			BirthYear:   intPtr(1765),
			DeathYear:   intPtr(1820),
			Description: "Vietnam's greatest poet, author of the epic poem 'Truyện Kiều' (The Tale of Kiều). His work represents the pinnacle of Vietnamese classical literature.",
			FamousQuote: "Trăm năm trong cõi người ta, chữ tài chữ mệnh khéo là ghét nhau",
			Achievements: jsonArray([]string{
				"Authored Truyện Kiều, masterpiece of Vietnamese literature",
				"Master of Nôm script poetry",
				"Influenced Vietnamese cultural identity",
				"UNESCO recognized his contribution to world literature",
			}),
			ImageURL:   "/assets/characters/nguyen_du.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ba_trieu",
			Name:        "Bà Triệu",
			Dynasty:     "Ngô",
			Rarity:      "Rare",
			BirthYear:   intPtr(225),
			DeathYear:   intPtr(248),
			Description: "Female warrior who led a rebellion against Wu Chinese occupation in the 3rd century. Known for her courage and determination in fighting foreign domination.",
			FamousQuote: "Tôi muốn cưỡi cơn gió mạnh, đạp sóng dữ, chém cá kình ở biển Đông",
			Achievements: jsonArray([]string{
				"Led rebellion against Chinese Wu dynasty",
				"Symbol of Vietnamese women's heroism",
				"Fought for Vietnamese independence in 3rd century",
				"Inspired future generations of patriots",
			}),
			ImageURL:   "/assets/characters/ba_trieu.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_vo_thi_sau",
			Name:        "Võ Thị Sáu",
			Dynasty:     "Cận đại",
			Rarity:      "Common",
			BirthYear:   intPtr(1933),
			DeathYear:   intPtr(1952),
			Description: "Young revolutionary who became a symbol of resistance during the French colonial period. Executed at age 19 for her anti-colonial activities.",
			FamousQuote: "Tôi chết vì Tổ quốc, lòng tôi vui sướng lắm",
			Achievements: jsonArray([]string{
				"Active member of resistance movement",
				"Symbol of young Vietnamese patriotism",
				"Sacrificed life for independence cause",
				"Inspired anti-colonial resistance",
			}),
			ImageURL:   "/assets/characters/vo_thi_sau.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ho_chi_minh",
			Name:        "Hồ Chí Minh",
			Dynasty:     "Cận đại",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1890),
			DeathYear:   intPtr(1969),
			Description: "Founding father of modern Vietnam and leader of independence movement. Led the country's struggle against French colonialism and American intervention.",
			FamousQuote: "Không có gì quý hơn độc lập tự do",
			Achievements: jsonArray([]string{
				"Founded Democratic Republic of Vietnam",
				"Led independence movement against France",
				"Declared Vietnamese independence (1945)",
				"Father of modern Vietnamese nation",
			}),
			ImageURL:   "/assets/characters/ho_chi_minh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	return characters
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func jsonArray(items []string) json.RawMessage {
	data, _ := json.Marshal(items)
	return json.RawMessage(data)
}

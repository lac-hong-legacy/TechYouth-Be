// seeders/lesson_seeder.go
package seeders

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"gorm.io/gorm"
)

// LessonSeeder handles seeding lessons for characters
type LessonSeeder struct {
	db *gorm.DB
}

// NewLessonSeeder creates a new lesson seeder
func NewLessonSeeder(db *gorm.DB) *LessonSeeder {
	return &LessonSeeder{db: db}
}

// SeedLessons seeds the database with lessons for historical characters
func (s *LessonSeeder) SeedLessons() error {
	lessons := s.getHistoricalLessons()

	for _, lesson := range lessons {
		// Check if lesson already exists
		var existingLesson model.Lesson
		if err := s.db.Where("id = ?", lesson.ID).First(&existingLesson).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Lesson doesn't exist, create it
				if err := s.db.Create(&lesson).Error; err != nil {
					log.Printf("Error creating lesson %s: %v", lesson.Title, err)
					return err
				}
				log.Printf("Created lesson: %s", lesson.Title)
			} else {
				log.Printf("Error checking lesson %s: %v", lesson.Title, err)
				return err
			}
		} else {
			log.Printf("Lesson %s already exists, skipping", lesson.Title)
		}
	}

	log.Println("Lesson seeding completed successfully")
	return nil
}

// getHistoricalLessons returns sample lessons for key historical characters
func (s *LessonSeeder) getHistoricalLessons() []model.Lesson {
	now := time.Now()

	lessons := []model.Lesson{
		// Hùng Vương lessons
		{
			ID:           "lesson_hung_vuong_1",
			CharacterID:  "char_hung_vuong",
			Title:        "The Birth of Văn Lang",
			Order:        1,
			Story:        "Long ago, in the misty mountains of northern Vietnam, lived the Dragon King Lạc Long Quân and the Fairy Mother Âu Cơ. From their union came 100 sons, and the eldest became the first Hùng King, founding the ancient kingdom of Văn Lang...",
			VoiceOverURL: "/assets/audio/hung_vuong_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv1_1",
					Type:     "multiple_choice",
					Question: "Who founded the first Vietnamese kingdom of Văn Lang?",
					Options:  []string{"Hùng Vương", "Lý Thái Tổ", "Ngô Quyền", "Đinh Bộ Lĩnh"},
					Answer:   "Hùng Vương",
					Points:   10,
				},
				{
					ID:       "q_hv1_2",
					Type:     "fill_blank",
					Question: "The ancient kingdom founded by Hùng Vương was called _____.",
					Answer:   "Văn Lang",
					Points:   15,
				},
			}),
			XPReward:  50,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_hung_vuong_2",
			CharacterID:  "char_hung_vuong",
			Title:        "The Lạc Việt People",
			Order:        2,
			Story:        "The Hùng Kings ruled over the Lạc Việt people, who were skilled in bronze-making, rice cultivation, and boat building. They created a sophisticated society along the Red River delta, laying the foundation for Vietnamese civilization...",
			VoiceOverURL: "/assets/audio/hung_vuong_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv2_1",
					Type:     "multiple_choice",
					Question: "The Hùng Kings ruled over which ancient Vietnamese people?",
					Options:  []string{"Cham people", "Lạc Việt", "Kinh people", "Hmong people"},
					Answer:   "Lạc Việt",
					Points:   10,
				},
			}),
			XPReward:  50,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Trần Hưng Đạo lessons
		{
			ID:           "lesson_tran_hung_dao_1",
			CharacterID:  "char_tran_hung_dao",
			Title:        "The Mongol Threat",
			Order:        1,
			Story:        "In the 13th century, the mighty Mongol Empire had conquered China and was turning its attention to Đại Việt. Kublai Khan's armies seemed unstoppable, but Prince Trần Quốc Tuấn, known as Trần Hưng Đạo, would prove them wrong...",
			VoiceOverURL: "/assets/audio/tran_hung_dao_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd1_1",
					Type:     "multiple_choice",
					Question: "Which empire threatened Đại Việt in the 13th century?",
					Options:  []string{"Chinese Ming", "Mongol Empire", "Khmer Empire", "Cham Kingdom"},
					Answer:   "Mongol Empire",
					Points:   10,
				},
				{
					ID:       "q_thd1_2",
					Type:     "multiple_choice",
					Question: "Who was the leader of the Mongol Empire during the invasions of Vietnam?",
					Options:  []string{"Genghis Khan", "Ögedei Khan", "Kublai Khan", "Möngke Khan"},
					Answer:   "Kublai Khan",
					Points:   15,
				},
			}),
			XPReward:  75,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_tran_hung_dao_2",
			CharacterID:  "char_tran_hung_dao",
			Title:        "Victory at Bạch Đằng",
			Order:        2,
			Story:        "Using ancient tactics and superior knowledge of local waters, Trần Hưng Đạo planted iron-tipped wooden stakes in the Bạch Đằng River. When the Mongol fleet attacked during high tide, the Vietnamese lured them forward, then retreated as the tide fell, leaving the enemy ships impaled...",
			VoiceOverURL: "/assets/audio/tran_hung_dao_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd2_1",
					Type:     "multiple_choice",
					Question: "What strategy did Trần Hưng Đạo use at Bạch Đằng River?",
					Options:  []string{"Direct naval battle", "Iron stakes and tide tactics", "Land siege", "Cavalry charge"},
					Answer:   "Iron stakes and tide tactics",
					Points:   15,
				},
				{
					ID:       "q_thd2_2",
					Type:     "drag_drop",
					Question: "Order the steps of the Bạch Đằng strategy:",
					Options:  []string{"Plant iron stakes", "Lure enemy fleet", "Wait for low tide", "Victory"},
					Answer:   []string{"Plant iron stakes", "Lure enemy fleet", "Wait for low tide", "Victory"},
					Points:   20,
				},
			}),
			XPReward:  100,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Lê Lợi lessons
		{
			ID:           "lesson_le_loi_1",
			CharacterID:  "char_le_loi",
			Title:        "The Ming Occupation",
			Order:        1,
			Story:        "After the fall of the Hồ dynasty, the Chinese Ming forces occupied Vietnam for 20 years. The people suffered under harsh rule and cultural suppression. But in the mountains of Thanh Hóa, a landlord named Lê Lợi began gathering patriots to resist...",
			VoiceOverURL: "/assets/audio/le_loi_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ll1_1",
					Type:     "multiple_choice",
					Question: "Which Chinese dynasty occupied Vietnam before Lê Lợi's rebellion?",
					Options:  []string{"Tang Dynasty", "Song Dynasty", "Ming Dynasty", "Qing Dynasty"},
					Answer:   "Ming Dynasty",
					Points:   10,
				},
				{
					ID:       "q_ll1_2",
					Type:     "fill_blank",
					Question: "Lê Lợi started his rebellion in the province of _____.",
					Answer:   "Thanh Hóa",
					Points:   15,
				},
			}),
			XPReward:  60,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_le_loi_2",
			CharacterID:  "char_le_loi",
			Title:        "The Legend of the Returned Sword",
			Order:        2,
			Story:        "Legend tells that Lê Lợi received a magical sword from the Dragon King to drive out the invaders. After victory, while boating on a lake in Thăng Long, a golden turtle surfaced and reclaimed the sword. The lake was thereafter called Hồ Hoàn Kiếm - Lake of the Returned Sword...",
			VoiceOverURL: "/assets/audio/le_loi_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ll2_1",
					Type:     "multiple_choice",
					Question: "What does 'Hồ Hoàn Kiếm' mean in English?",
					Options:  []string{"Dragon Lake", "Victory Lake", "Lake of the Returned Sword", "Golden Turtle Lake"},
					Answer:   "Lake of the Returned Sword",
					Points:   10,
				},
				{
					ID:       "q_ll2_2",
					Type:     "multiple_choice",
					Question: "Who gave Lê Lợi the magical sword according to legend?",
					Options:  []string{"Buddha", "Dragon King", "Jade Emperor", "Mountain Spirit"},
					Answer:   "Dragon King",
					Points:   15,
				},
			}),
			XPReward:  75,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Hai Bà Trưng lessons
		{
			ID:           "lesson_hai_ba_trung_1",
			CharacterID:  "char_hai_ba_trung",
			Title:        "Sisters of Resistance",
			Order:        1,
			Story:        "In 40 AD, when Chinese Han officials became increasingly oppressive, two noble sisters from Mê Linh decided to act. Trưng Trắc and Trưng Nhị, trained in martial arts and military strategy, could no longer watch their people suffer...",
			VoiceOverURL: "/assets/audio/hai_ba_trung_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hbt1_1",
					Type:     "multiple_choice",
					Question: "The Trưng Sisters led their rebellion in which year?",
					Options:  []string{"30 AD", "40 AD", "50 AD", "60 AD"},
					Answer:   "40 AD",
					Points:   10,
				},
				{
					ID:       "q_hbt1_2",
					Type:     "fill_blank",
					Question: "The Trưng Sisters were from _____ district.",
					Answer:   "Mê Linh",
					Points:   15,
				},
			}),
			XPReward:  65,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Ngô Quyền lessons
		{
			ID:           "lesson_ngo_quyen_1",
			CharacterID:  "char_ngo_quyen",
			Title:        "The End of a Thousand Years",
			Order:        1,
			Story:        "For nearly a thousand years, Vietnam had been under Chinese rule. But in 938 AD, a military commander named Ngô Quyền would change the course of history forever at the Bạch Đằng River...",
			VoiceOverURL: "/assets/audio/ngo_quyen_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_nq1_1",
					Type:     "multiple_choice",
					Question: "How long had Vietnam been under Chinese rule before Ngô Quyền's victory?",
					Options:  []string{"500 years", "800 years", "Nearly 1000 years", "1200 years"},
					Answer:   "Nearly 1000 years",
					Points:   15,
				},
				{
					ID:       "q_nq1_2",
					Type:     "multiple_choice",
					Question: "Where did Ngô Quyền win his decisive victory?",
					Options:  []string{"Red River", "Bạch Đằng River", "Perfume River", "Mekong River"},
					Answer:   "Bạch Đằng River",
					Points:   10,
				},
			}),
			XPReward:  80,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Lý Thái Tổ lessons
		{
			ID:           "lesson_ly_thai_to_1",
			CharacterID:  "char_ly_thai_to",
			Title:        "The Move to Thăng Long",
			Order:        1,
			Story:        "In 1010, Emperor Lý Thái Tổ made a momentous decision - to move the capital from Hoa Lư to the site of modern-day Hanoi. He named it Thăng Long, meaning 'Rising Dragon', after witnessing a golden dragon ascending from the Red River...",
			VoiceOverURL: "/assets/audio/ly_thai_to_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ltt1_1",
					Type:     "multiple_choice",
					Question: "What does 'Thăng Long' mean?",
					Options:  []string{"Golden City", "Rising Dragon", "Royal Capital", "Sacred Land"},
					Answer:   "Rising Dragon",
					Points:   10,
				},
				{
					ID:       "q_ltt1_2",
					Type:     "multiple_choice",
					Question: "When did Lý Thái Tổ move the capital to Thăng Long?",
					Options:  []string{"1009", "1010", "1011", "1012"},
					Answer:   "1010",
					Points:   15,
				},
			}),
			XPReward:  70,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Hồ Chí Minh lessons
		{
			ID:           "lesson_ho_chi_minh_1",
			CharacterID:  "char_ho_chi_minh",
			Title:        "The Young Patriot",
			Order:        1,
			Story:        "Born as Nguyễn Sinh Cung in 1890, the future leader of Vietnam grew up witnessing French colonial oppression. As a young man, he left Vietnam on a French steamship, beginning a journey that would take him around the world and shape his revolutionary ideals...",
			VoiceOverURL: "/assets/audio/ho_chi_minh_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hcm1_1",
					Type:     "multiple_choice",
					Question: "What was Hồ Chí Minh's birth name?",
					Options:  []string{"Nguyễn Tất Thành", "Nguyễn Sinh Cung", "Nguyễn Ái Quốc", "Phan Bội Châu"},
					Answer:   "Nguyễn Sinh Cung",
					Points:   15,
				},
				{
					ID:       "q_hcm1_2",
					Type:     "fill_blank",
					Question: "Hồ Chí Minh was born in the year _____.",
					Answer:   "1890",
					Points:   10,
				},
			}),
			XPReward:  85,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_ho_chi_minh_2",
			CharacterID:  "char_ho_chi_minh",
			Title:        "Declaration of Independence",
			Order:        2,
			Story:        "On September 2, 1945, in Ba Đình Square, Hồ Chí Minh read the Declaration of Independence, establishing the Democratic Republic of Vietnam. His opening words quoted the American Declaration of Independence: 'All men are created equal...'",
			VoiceOverURL: "/assets/audio/ho_chi_minh_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hcm2_1",
					Type:     "multiple_choice",
					Question: "When did Hồ Chí Minh declare Vietnamese independence?",
					Options:  []string{"August 19, 1945", "September 2, 1945", "October 10, 1945", "December 19, 1946"},
					Answer:   "September 2, 1945",
					Points:   15,
				},
				{
					ID:       "q_hcm2_2",
					Type:     "multiple_choice",
					Question: "Where did Hồ Chí Minh read the Declaration of Independence?",
					Options:  []string{"Hoan Kiem Lake", "Ba Đình Square", "Presidential Palace", "National Assembly"},
					Answer:   "Ba Đình Square",
					Points:   10,
				},
			}),
			XPReward:  100,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	return lessons
}

// QuestionData represents question data for easier creation
type QuestionData struct {
	ID       string
	Type     string
	Question string
	Options  []string
	Answer   interface{}
	Points   int
	Metadata map[string]interface{}
}

// createQuestions converts QuestionData to JSON format
func createQuestions(questionsData []QuestionData) json.RawMessage {
	var questions []model.Question

	for _, qData := range questionsData {
		question := model.Question{
			ID:       qData.ID,
			Type:     qData.Type,
			Question: qData.Question,
			Options:  qData.Options,
			Answer:   qData.Answer,
			Points:   qData.Points,
			Metadata: qData.Metadata,
		}
		questions = append(questions, question)
	}

	data, _ := json.Marshal(questions)
	return json.RawMessage(data)
}

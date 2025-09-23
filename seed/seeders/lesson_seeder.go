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
			Title:        "Sự ra đời của Văn Lang",
			Order:        1,
			Story:        "Ngày xưa, trên những núi non sương mù của miền Bắc Việt Nam, sống Vua Rồng Lạc Long Quân và Mẹ Tiên Âu Cơ. Từ sự kết hợp của họ sinh ra 100 người con trai, và người con cả trở thành vua Hùng đầu tiên, sáng lập vương quốc cổ xưa Văn Lang...",
			VoiceOverURL: "/assets/audio/hung_vuong_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv1_1",
					Type:     "multiple_choice",
					Question: "Ai đã sáng lập vương quốc Việt Nam đầu tiên là Văn Lang?",
					Options:  []string{"Hùng Vương", "Lý Thái Tổ", "Ngô Quyền", "Đinh Bộ Lĩnh"},
					Answer:   "Hùng Vương",
					Points:   10,
				},
				{
					ID:       "q_hv1_2",
					Type:     "fill_blank",
					Question: "Vương quốc cổ xưa được Hùng Vương sáng lập có tên là _____.",
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
			Title:        "Dân tộc Lạc Việt",
			Order:        2,
			Story:        "Các vua Hùng cai trị dân tộc Lạc Việt, những người thành thạo trong việc chế tạo đồng, trồng lúa và đóng thuyền. Họ đã tạo ra một xã hội tinh vi dọc theo vùng đồng bằng sông Hồng, đặt nền tảng cho nền văn minh Việt Nam...",
			VoiceOverURL: "/assets/audio/hung_vuong_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv2_1",
					Type:     "multiple_choice",
					Question: "Các vua Hùng cai trị dân tộc Việt Nam cổ đại nào?",
					Options:  []string{"Dân tộc Chăm", "Lạc Việt", "Dân tộc Kinh", "Dân tộc Hmông"},
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
			Title:        "Mối đe dọa từ quân Mông Cổ",
			Order:        1,
			Story:        "Vào thế kỷ 13, đế chế Mông Cổ hùng mạnh đã chinh phục Trung Quốc và đang hướng sự chú ý đến Đại Việt. Quân đội của Hốt Tất Liệt dường như không thể ngăn cản, nhưng Hoàng tử Trần Quốc Tuấn, được biết đến với tên Trần Hưng Đạo, sẽ chứng minh họ sai lầm...",
			VoiceOverURL: "/assets/audio/tran_hung_dao_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd1_1",
					Type:     "multiple_choice",
					Question: "Đế chế nào đã đe dọa Đại Việt trong thế kỷ 13?",
					Options:  []string{"Nhà Minh Trung Quốc", "Đế chế Mông Cổ", "Đế chế Khmer", "Vương quốc Chăm"},
					Answer:   "Đế chế Mông Cổ",
					Points:   10,
				},
				{
					ID:       "q_thd1_2",
					Type:     "multiple_choice",
					Question: "Ai là lãnh tụ của Đế chế Mông Cổ trong các cuộc xâm lược Việt Nam?",
					Options:  []string{"Thành Cát Tư Hãn", "Oa Khoát Đài", "Hốt Tất Liệt", "Mông Cát"},
					Answer:   "Hốt Tất Liệt",
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
			Title:        "Chiến thắng tại Bạch Đằng",
			Order:        2,
			Story:        "Sử dụng chiến thuật cổ xưa và kiến thức ưu việt về vùng nước địa phương, Trần Hưng Đạo đã cắm những cọc gỗ có đầu nhọn bằng sắt trong sông Bạch Đằng. Khi hạm đội Mông Cổ tấn công trong thủy triều cao, người Việt dụ họ tiến về phía trước, sau đó rút lui khi thủy triều xuống, để lại các tàu địch bị đâm thủng...",
			VoiceOverURL: "/assets/audio/tran_hung_dao_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd2_1",
					Type:     "multiple_choice",
					Question: "Trần Hưng Đạo đã sử dụng chiến lược gì tại sông Bạch Đằng?",
					Options:  []string{"Trận chiến hải quân trực tiếp", "Cọc sắt và chiến thuật thủy triều", "Bao vây trên đất liền", "Tấn công kỵ binh"},
					Answer:   "Cọc sắt và chiến thuật thủy triều",
					Points:   15,
				},
				{
					ID:       "q_thd2_2",
					Type:     "drag_drop",
					Question: "Sắp xếp các bước của chiến lược Bạch Đằng:",
					Options:  []string{"Cắm cọc sắt", "Dụ hạm đội địch", "Chờ thủy triều xuống", "Chiến thắng"},
					Answer:   []string{"Cắm cọc sắt", "Dụ hạm đội địch", "Chờ thủy triều xuống", "Chiến thắng"},
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
			Title:        "Sự chiếm đóng của nhà Minh",
			Order:        1,
			Story:        "Sau sự sụp đổ của triều đại Hồ, quân Minh Trung Quốc đã chiếm đóng Việt Nam trong 20 năm. Người dân phải chịu đựng sự cai trị khắc nghiệt và áp chế văn hóa. Nhưng trên những ngọn núi Thanh Hóa, một địa chủ tên Lê Lợi bắt đầu tập hợp những người yêu nước để kháng chiến...",
			VoiceOverURL: "/assets/audio/le_loi_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ll1_1",
					Type:     "multiple_choice",
					Question: "Triều đại Trung Quốc nào đã chiếm đóng Việt Nam trước cuộc nổi dậy của Lê Lợi?",
					Options:  []string{"Triều đại Đường", "Triều đại Tống", "Triều đại Minh", "Triều đại Thanh"},
					Answer:   "Triều đại Minh",
					Points:   10,
				},
				{
					ID:       "q_ll1_2",
					Type:     "fill_blank",
					Question: "Lê Lợi bắt đầu cuộc nổi dậy ở tỉnh _____.",
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
			Title:        "Truyền thuyết Hoàn Kiếm",
			Order:        2,
			Story:        "Truyền thuyết kể rằng Lê Lợi đã nhận được một thanh kiếm thần từ Vua Rồng để đuổi quân xâm lược. Sau chiến thắng, khi đang chèo thuyền trên một hồ ở Thăng Long, một con rùa vàng nổi lên mặt nước và đòi lại thanh kiếm. Từ đó hồ được gọi là Hồ Hoàn Kiếm...",
			VoiceOverURL: "/assets/audio/le_loi_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ll2_1",
					Type:     "multiple_choice",
					Question: "'Hồ Hoàn Kiếm' có nghĩa là gì?",
					Options:  []string{"Hồ Rồng", "Hồ Chiến thắng", "Hồ Hoàn Kiếm", "Hồ Rùa Vàng"},
					Answer:   "Hồ Hoàn Kiếm",
					Points:   10,
				},
				{
					ID:       "q_ll2_2",
					Type:     "multiple_choice",
					Question: "Theo truyền thuyết, ai đã tặng Lê Lợi thanh kiếm thần?",
					Options:  []string{"Phật", "Vua Rồng", "Ngọc Hoàng", "Thần Núi"},
					Answer:   "Vua Rồng",
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
			Title:        "Hai chị em kháng chiến",
			Order:        1,
			Story:        "Năm 40 sau Công nguyên, khi các quan lại Hán Trung Quốc ngày càng áp bức, hai chị em quý tộc từ Mê Linh quyết định hành động. Trưng Trắc và Trưng Nhị, được huấn luyện võ thuật và chiến lược quân sự, không thể tiếp tục nhìn người dân của mình chịu đựng...",
			VoiceOverURL: "/assets/audio/hai_ba_trung_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hbt1_1",
					Type:     "multiple_choice",
					Question: "Hai Bà Trưng lãnh đạo cuộc nổi dậy vào năm nào?",
					Options:  []string{"30 SCN", "40 SCN", "50 SCN", "60 SCN"},
					Answer:   "40 SCN",
					Points:   10,
				},
				{
					ID:       "q_hbt1_2",
					Type:     "fill_blank",
					Question: "Hai Bà Trưng đến từ huyện _____.",
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
			Title:        "Kết thúc một thiên niên kỷ",
			Order:        1,
			Story:        "Trong gần một thiên niên kỷ, Việt Nam đã nằm dưới sự cai trị của Trung Quốc. Nhưng vào năm 938 sau Công nguyên, một tướng lĩnh tên Ngô Quyền sẽ thay đổi tiến trình lịch sử mãi mãi tại sông Bạch Đằng...",
			VoiceOverURL: "/assets/audio/ngo_quyen_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_nq1_1",
					Type:     "multiple_choice",
					Question: "Việt Nam đã nằm dưới sự cai trị của Trung Quốc trong bao lâu trước chiến thắng của Ngô Quyền?",
					Options:  []string{"500 năm", "800 năm", "Gần 1000 năm", "1200 năm"},
					Answer:   "Gần 1000 năm",
					Points:   15,
				},
				{
					ID:       "q_nq1_2",
					Type:     "multiple_choice",
					Question: "Ngô Quyền đã giành chiến thắng quyết định ở đâu?",
					Options:  []string{"Sông Hồng", "Sông Bạch Đằng", "Sông Hương", "Sông Mê Kông"},
					Answer:   "Sông Bạch Đằng",
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
			Title:        "Việc dời đô về Thăng Long",
			Order:        1,
			Story:        "Năm 1010, Hoàng đế Lý Thái Tổ đã đưa ra một quyết định trọng đại - di chuyển kinh đô từ Hoa Lư đến vị trí của Hà Nội ngày nay. Ngài đặt tên là Thăng Long, có nghĩa là 'Rồng Bay Lên', sau khi chứng kiến một con rồng vàng bay lên từ sông Hồng...",
			VoiceOverURL: "/assets/audio/ly_thai_to_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ltt1_1",
					Type:     "multiple_choice",
					Question: "'Thăng Long' có nghĩa là gì?",
					Options:  []string{"Thành Phố Vàng", "Rồng Bay Lên", "Kinh Đô Hoàng Gia", "Đất Thiêng"},
					Answer:   "Rồng Bay Lên",
					Points:   10,
				},
				{
					ID:       "q_ltt1_2",
					Type:     "multiple_choice",
					Question: "Lý Thái Tổ dời đô về Thăng Long vào năm nào?",
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
			Title:        "Người thanh niên yêu nước",
			Order:        1,
			Story:        "Sinh ra với tên Nguyễn Sinh Cung vào năm 1890, vị lãnh tụ tương lai của Việt Nam lớn lên trong việc chứng kiến sự áp bức thuộc địa của Pháp. Khi còn trẻ, ông rời Việt Nam trên một tàu hơi nước của Pháp, bắt đầu cuộc hành trình đưa ông đi khắp thế giới và hình thành lý tưởng cách mạng...",
			VoiceOverURL: "/assets/audio/ho_chi_minh_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hcm1_1",
					Type:     "multiple_choice",
					Question: "Tên khai sinh của Hồ Chí Minh là gì?",
					Options:  []string{"Nguyễn Tất Thành", "Nguyễn Sinh Cung", "Nguyễn Ái Quốc", "Phan Bội Châu"},
					Answer:   "Nguyễn Sinh Cung",
					Points:   15,
				},
				{
					ID:       "q_hcm1_2",
					Type:     "fill_blank",
					Question: "Hồ Chí Minh sinh năm _____.",
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
			Title:        "Tuyên ngôn Độc lập",
			Order:        2,
			Story:        "Ngày 2 tháng 9 năm 1945, tại Quảng trường Ba Đình, Hồ Chí Minh đã đọc Tuyên ngôn Độc lập, thành lập nước Việt Nam Dân chủ Cộng hòa. Lời mở đầu của ông trích dẫn Tuyên ngôn Độc lập của Mỹ: 'Tất cả mọi người sinh ra đều bình đẳng...'",
			VoiceOverURL: "/assets/audio/ho_chi_minh_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hcm2_1",
					Type:     "multiple_choice",
					Question: "Hồ Chí Minh tuyên bố độc lập của Việt Nam vào ngày nào?",
					Options:  []string{"19 tháng 8, 1945", "2 tháng 9, 1945", "10 tháng 10, 1945", "19 tháng 12, 1946"},
					Answer:   "2 tháng 9, 1945",
					Points:   15,
				},
				{
					ID:       "q_hcm2_2",
					Type:     "multiple_choice",
					Question: "Hồ Chí Minh đọc Tuyên ngôn Độc lập ở đâu?",
					Options:  []string{"Hồ Hoàn Kiếm", "Quảng trường Ba Đình", "Phủ Chủ tịch", "Quốc hội"},
					Answer:   "Quảng trường Ba Đình",
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

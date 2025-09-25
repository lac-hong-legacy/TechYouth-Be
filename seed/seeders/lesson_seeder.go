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
			ID:          "lesson_hung_vuong_1",
			CharacterID: "char_hung_vuong",
			Title:       "Sự ra đời của Văn Lang",
			Order:       1,
			Story:       "Ngày xưa, trên những núi non sương mù của miền Bắc Việt Nam, sống Vua Rồng Lạc Long Quân và Mẹ Tiên Âu Cơ. Từ sự kết hợp của họ sinh ra 100 người con trai, và người con cả trở thành vua Hùng đầu tiên, sáng lập vương quốc cổ xưa Văn Lang...",

			// Media Content (to be uploaded via Admin API)
			VideoURL:      "", // Will be set when video is uploaded via admin
			SubtitleURL:   "", // Will be set when subtitle is uploaded via admin
			ThumbnailURL:  "", // Will be auto-generated or uploaded via admin
			VideoDuration: 0,  // Will be set automatically when video is processed
			CanSkipAfter:  5,
			HasSubtitles:  false, // Will be set to true when subtitle is uploaded
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
				{
					ID:       "q_hv1_3",
					Type:     "multiple_choice",
					Question: "Theo truyền thuyết, Hùng Vương là con của ai?",
					Options:  []string{"Lạc Long Quân và Âu Cơ", "Thần Nông và Tiên Nữ", "Vua Rồng và Công chúa", "Thần Sấm và Mẹ Nước"},
					Answer:   "Lạc Long Quân và Âu Cơ",
					Points:   15,
				},
				// {
				// 	ID:       "q_hv1_4",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự thời gian các sự kiện trong truyền thuyết:",
				// 	Options:  []string{"Lạc Long Quân gặp Âu Cơ", "Sinh ra 100 người con", "Chia làm hai nhóm", "Thành lập Văn Lang"},
				// 	Answer:   []string{"Lạc Long Quân gặp Âu Cơ", "Sinh ra 100 người con", "Chia làm hai nhóm", "Thành lập Văn Lang"},
				// 	Points:   20,
				// },
				{
					ID:       "q_hv1_5",
					Type:     "connect",
					Question: "Nối các khái niệm với ý nghĩa của chúng:",
					Options:  []string{"Văn Lang", "Hùng Vương", "Lạc Long Quân", "Âu Cơ"},
					Answer:   map[string]string{"Văn Lang": "Vương quốc đầu tiên", "Hùng Vương": "Vua đầu tiên", "Lạc Long Quân": "Vua Rồng", "Âu Cơ": "Mẹ Tiên"},
					Points:   25,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Văn Lang": "Vương quốc đầu tiên", "Hùng Vương": "Vua đầu tiên", "Lạc Long Quân": "Vua Rồng", "Âu Cơ": "Mẹ Tiên"}},
				},
			}),
			XPReward:  50,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "lesson_hung_vuong_2",
			CharacterID: "char_hung_vuong",
			Title:       "Dân tộc Lạc Việt",
			Order:       2,
			Story:       "Các vua Hùng cai trị dân tộc Lạc Việt, những người thành thạo trong việc chế tạo đồng, trồng lúa và đóng thuyền. Họ đã tạo ra một xã hội tinh vi dọc theo vùng đồng bằng sông Hồng, đặt nền tảng cho nền văn minh Việt Nam...",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv2_1",
					Type:     "multiple_choice",
					Question: "Các vua Hùng cai trị dân tộc Việt Nam cổ đại nào?",
					Options:  []string{"Dân tộc Chăm", "Lạc Việt", "Dân tộc Kinh", "Dân tộc Hmông"},
					Answer:   "Lạc Việt",
					Points:   10,
				},
				{
					ID:       "q_hv2_2",
					Type:     "multiple_choice",
					Question: "Dân tộc Lạc Việt nổi tiếng với kỹ năng gì?",
					Options:  []string{"Chế tạo đồng và trồng lúa", "Chăn nuôi gia súc", "Đánh cá trên biển", "Thương mại với Trung Quốc"},
					Answer:   "Chế tạo đồng và trồng lúa",
					Points:   15,
				},
				{
					ID:       "q_hv2_3",
					Type:     "fill_blank",
					Question: "Dân tộc Lạc Việt sống chủ yếu ở vùng đồng bằng sông _____.",
					Answer:   "Hồng",
					Points:   10,
				},
				{
					ID:       "q_hv2_4",
					Type:     "multiple_choice",
					Question: "Ngoài chế tạo đồng và trồng lúa, dân tộc Lạc Việt còn thành thạo việc gì?",
					Options:  []string{"Dệt lụa", "Đóng thuyền", "Làm gốm", "Chạm khắc đá"},
					Answer:   "Đóng thuyền",
					Points:   15,
				},
				{
					ID:       "q_hv2_5",
					Type:     "connect",
					Question: "Nối các hoạt động với vùng địa lý phù hợp:",
					Options:  []string{"Trồng lúa", "Chế tạo đồng", "Đóng thuyền", "Xây dựng làng"},
					Answer:   map[string]string{"Trồng lúa": "Đồng bằng", "Chế tạo đồng": "Vùng có quặng", "Đóng thuyền": "Ven sông", "Xây dựng làng": "Đất cao"},
					Points:   20,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Trồng lúa": "Đồng bằng", "Chế tạo đồng": "Vùng có quặng", "Đóng thuyền": "Ven sông", "Xây dựng làng": "Đất cao"}},
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
			ID:          "lesson_tran_hung_dao_1",
			CharacterID: "char_tran_hung_dao",
			Title:       "Mối đe dọa từ quân Mông Cổ",
			Order:       1,
			Story:       "Vào thế kỷ 13, đế chế Mông Cổ hùng mạnh đã chinh phục Trung Quốc và đang hướng sự chú ý đến Đại Việt. Quân đội của Hốt Tất Liệt dường như không thể ngăn cản, nhưng Hoàng tử Trần Quốc Tuấn, được biết đến với tên Trần Hưng Đạo, sẽ chứng minh họ sai lầm...",
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
				{
					ID:       "q_thd1_3",
					Type:     "fill_blank",
					Question: "Trần Quốc Tuấn được biết đến với tên gọi _____.",
					Answer:   "Trần Hưng Đạo",
					Points:   10,
				},
				{
					ID:       "q_thd1_4",
					Type:     "multiple_choice",
					Question: "Thế kỷ 13 là thời gian của triều đại nào ở Việt Nam?",
					Options:  []string{"Triều đại Lý", "Triều đại Trần", "Triều đại Lê", "Triều đại Nguyễn"},
					Answer:   "Triều đại Trần",
					Points:   15,
				},
				// {
				// 	ID:       "q_thd1_5",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự các sự kiện dẫn đến cuộc xâm lược Mông Cổ:",
				// 	Options:  []string{"Mông Cổ chinh phục Trung Quốc", "Hốt Tất Liệt lên kế hoạch", "Đe dọa Đại Việt", "Trần Hưng Đạo chuẩn bị"},
				// 	Answer:   []string{"Mông Cổ chinh phục Trung Quốc", "Hốt Tất Liệt lên kế hoạch", "Đe dọa Đại Việt", "Trần Hưng Đạo chuẩn bị"},
				// 	Points:   20,
				// },
			}),
			XPReward:  75,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "lesson_tran_hung_dao_2",
			CharacterID: "char_tran_hung_dao",
			Title:       "Chiến thắng tại Bạch Đằng",
			Order:       2,
			Story:       "Sử dụng chiến thuật cổ xưa và kiến thức ưu việt về vùng nước địa phương, Trần Hưng Đạo đã cắm những cọc gỗ có đầu nhọn bằng sắt trong sông Bạch Đằng. Khi hạm đội Mông Cổ tấn công trong thủy triều cao, người Việt dụ họ tiến về phía trước, sau đó rút lui khi thủy triều xuống, để lại các tàu địch bị đâm thủng...",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd2_1",
					Type:     "multiple_choice",
					Question: "Trần Hưng Đạo đã sử dụng chiến lược gì tại sông Bạch Đằng?",
					Options:  []string{"Trận chiến hải quân trực tiếp", "Cọc sắt và chiến thuật thủy triều", "Bao vây trên đất liền", "Tấn công kỵ binh"},
					Answer:   "Cọc sắt và chiến thuật thủy triều",
					Points:   15,
				},
				// {
				// 	ID:       "q_thd2_2",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp các bước của chiến lược Bạch Đằng:",
				// 	Options:  []string{"Cắm cọc sắt", "Dụ hạm đội địch", "Chờ thủy triều xuống", "Chiến thắng"},
				// 	Answer:   []string{"Cắm cọc sắt", "Dụ hạm đội địch", "Chờ thủy triều xuống", "Chiến thắng"},
				// 	Points:   20,
				// },
				{
					ID:       "q_thd2_3",
					Type:     "multiple_choice",
					Question: "Tại sao Trần Hưng Đạo chọn sông Bạch Đằng làm địa điểm chiến đấu?",
					Options:  []string{"Gần kinh đô", "Có thủy triều lên xuống", "Nước sâu nhất", "Địch không biết đường"},
					Answer:   "Có thủy triều lên xuống",
					Points:   15,
				},
				{
					ID:       "q_thd2_4",
					Type:     "fill_blank",
					Question: "Các cọc gỗ được cắm trong sông có đầu nhọn bằng _____.",
					Answer:   "sắt",
					Points:   10,
				},
				{
					ID:       "q_thd2_5",
					Type:     "multiple_choice",
					Question: "Chiến thuật của Trần Hưng Đạo thể hiện điều gì?",
					Options:  []string{"Sức mạnh quân sự", "Trí tuệ và hiểu biết địa lý", "Vũ khí hiện đại", "Quân số đông đảo"},
					Answer:   "Trí tuệ và hiểu biết địa lý",
					Points:   20,
				},
				{
					ID:       "q_thd2_6",
					Type:     "connect",
					Question: "Nối các yếu tố với vai trò trong chiến thắng:",
					Options:  []string{"Cọc sắt", "Thủy triều", "Hạm đội địch", "Quân Việt"},
					Answer:   map[string]string{"Cọc sắt": "Phá hủy tàu", "Thủy triều": "Tạo bẫy tự nhiên", "Hạm đội địch": "Rơi vào bẫy", "Quân Việt": "Dụ địch tiến"},
					Points:   25,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Cọc sắt": "Phá hủy tàu", "Thủy triều": "Tạo bẫy tự nhiên", "Hạm đội địch": "Rơi vào bẫy", "Quân Việt": "Dụ địch tiến"}},
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
			ID:          "lesson_le_loi_1",
			CharacterID: "char_le_loi",
			Title:       "Sự chiếm đóng của nhà Minh",
			Order:       1,
			Story:       "Sau sự sụp đổ của triều đại Hồ, quân Minh Trung Quốc đã chiếm đóng Việt Nam trong 20 năm. Người dân phải chịu đựng sự cai trị khắc nghiệt và áp chế văn hóa. Nhưng trên những ngọn núi Thanh Hóa, một địa chủ tên Lê Lợi bắt đầu tập hợp những người yêu nước để kháng chiến...",
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
				{
					ID:       "q_ll1_3",
					Type:     "multiple_choice",
					Question: "Quân Minh chiếm đóng Việt Nam trong bao lâu?",
					Options:  []string{"10 năm", "15 năm", "20 năm", "25 năm"},
					Answer:   "20 năm",
					Points:   15,
				},
				{
					ID:       "q_ll1_4",
					Type:     "multiple_choice",
					Question: "Triều đại nào đã sụp đổ trước khi quân Minh chiếm đóng?",
					Options:  []string{"Triều đại Trần", "Triều đại Hồ", "Triều đại Lý", "Triều đại Lê"},
					Answer:   "Triều đại Hồ",
					Points:   15,
				},
				{
					ID:       "q_ll1_5",
					Type:     "multiple_choice",
					Question: "Lê Lợi có xuất thân từ tầng lớp nào?",
					Options:  []string{"Nông dân", "Địa chủ", "Quan lại", "Thương gia"},
					Answer:   "Địa chủ",
					Points:   10,
				},
				// {
				// 	ID:       "q_ll1_6",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự thời gian các sự kiện:",
				// 	Options:  []string{"Triều đại Hồ sụp đổ", "Quân Minh chiếm đóng", "Áp chế văn hóa", "Lê Lợi tập hợp người yêu nước"},
				// 	Answer:   []string{"Triều đại Hồ sụp đổ", "Quân Minh chiếm đóng", "Áp chế văn hóa", "Lê Lợi tập hợp người yêu nước"},
				// 	Points:   20,
				// },
			}),
			XPReward:  60,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "lesson_le_loi_2",
			CharacterID: "char_le_loi",
			Title:       "Truyền thuyết Hoàn Kiếm",
			Order:       2,
			Story:       "Truyền thuyết kể rằng Lê Lợi đã nhận được một thanh kiếm thần từ Vua Rồng để đuổi quân xâm lược. Sau chiến thắng, khi đang chèo thuyền trên một hồ ở Thăng Long, một con rùa vàng nổi lên mặt nước và đòi lại thanh kiếm. Từ đó hồ được gọi là Hồ Hoàn Kiếm...",
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
				{
					ID:       "q_ll2_3",
					Type:     "multiple_choice",
					Question: "Con vật nào đã đòi lại thanh kiếm từ Lê Lợi?",
					Options:  []string{"Rồng vàng", "Rùa vàng", "Phượng hoàng", "Kỳ lân"},
					Answer:   "Rùa vàng",
					Points:   15,
				},
				{
					ID:       "q_ll2_4",
					Type:     "fill_blank",
					Question: "Sự kiện hoàn trả kiếm diễn ra ở thành phố _____.",
					Answer:   "Thăng Long",
					Points:   10,
				},
				// {
				// 	ID:       "q_ll2_5",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự các sự kiện trong truyền thuyết:",
				// 	Options:  []string{"Nhận kiếm thần", "Đuổi quân xâm lược", "Chèo thuyền trên hồ", "Rùa vàng đòi lại kiếm"},
				// 	Answer:   []string{"Nhận kiếm thần", "Đuổi quân xâm lược", "Chèo thuyền trên hồ", "Rùa vàng đòi lại kiếm"},
				// 	Points:   20,
				// },
				{
					ID:       "q_ll2_6",
					Type:     "connect",
					Question: "Nối các yếu tố với ý nghĩa trong truyền thuyết:",
					Options:  []string{"Kiếm thần", "Vua Rồng", "Rùa vàng", "Hồ Hoàn Kiếm"},
					Answer:   map[string]string{"Kiếm thần": "Sức mạnh thiêng liêng", "Vua Rồng": "Người ban phước", "Rùa vàng": "Sứ giả thu hồi", "Hồ Hoàn Kiếm": "Nơi trả lại thiêng liêng"},
					Points:   25,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Kiếm thần": "Sức mạnh thiêng liêng", "Vua Rồng": "Người ban phước", "Rùa vàng": "Sứ giả thu hồi", "Hồ Hoàn Kiếm": "Nơi trả lại thiêng liêng"}},
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
			ID:          "lesson_hai_ba_trung_1",
			CharacterID: "char_hai_ba_trung",
			Title:       "Hai chị em kháng chiến",
			Order:       1,
			Story:       "Năm 40 sau Công nguyên, khi các quan lại Hán Trung Quốc ngày càng áp bức, hai chị em quý tộc từ Mê Linh quyết định hành động. Trưng Trắc và Trưng Nhị, được huấn luyện võ thuật và chiến lược quân sự, không thể tiếp tục nhìn người dân của mình chịu đựng...",
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
				{
					ID:       "q_hbt1_3",
					Type:     "multiple_choice",
					Question: "Tên của hai chị em Trưng là gì?",
					Options:  []string{"Trưng Trắc và Trưng Nhị", "Trưng Vương và Trưng Nữ", "Trưng Tráng và Trưng Nhi", "Trưng Trinh và Trưng Nương"},
					Answer:   "Trưng Trắc và Trưng Nhị",
					Points:   15,
				},
				{
					ID:       "q_hbt1_4",
					Type:     "multiple_choice",
					Question: "Hai Bà Trưng nổi dậy chống lại sự cai trị của ai?",
					Options:  []string{"Quan lại Hán Trung Quốc", "Vua Chăm", "Quân Mông Cổ", "Thực dân Pháp"},
					Answer:   "Quan lại Hán Trung Quốc",
					Points:   15,
				},
				{
					ID:       "q_hbt1_5",
					Type:     "multiple_choice",
					Question: "Hai Bà Trưng có xuất thân từ tầng lớp nào?",
					Options:  []string{"Nông dân", "Quý tộc", "Thương gia", "Quan lại"},
					Answer:   "Quý tộc",
					Points:   10,
				},
				{
					ID:       "q_hbt1_6",
					Type:     "multiple_choice",
					Question: "Hai Bà Trưng được huấn luyện về gì từ nhỏ?",
					Options:  []string{"Thơ văn", "Võ thuật và chiến lược quân sự", "Âm nhạc", "Hội họa"},
					Answer:   "Võ thuật và chiến lược quân sự",
					Points:   20,
				},
				{
					ID:       "q_hbt1_7",
					Type:     "connect",
					Question: "Nối các yếu tố với đặc điểm của Hai Bà Trưng:",
					Options:  []string{"Xuất thân", "Kỹ năng", "Động lực", "Thời gian"},
					Answer:   map[string]string{"Xuất thân": "Quý tộc Mê Linh", "Kỹ năng": "Võ thuật và chiến lược", "Động lực": "Chống áp bức", "Thời gian": "Năm 40 SCN"},
					Points:   25,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Xuất thân": "Quý tộc Mê Linh", "Kỹ năng": "Võ thuật và chiến lược", "Động lực": "Chống áp bức", "Thời gian": "Năm 40 SCN"}},
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
			ID:          "lesson_ngo_quyen_1",
			CharacterID: "char_ngo_quyen",
			Title:       "Kết thúc một thiên niên kỷ",
			Order:       1,
			Story:       "Trong gần một thiên niên kỷ, Việt Nam đã nằm dưới sự cai trị của Trung Quốc. Nhưng vào năm 938 sau Công nguyên, một tướng lĩnh tên Ngô Quyền sẽ thay đổi tiến trình lịch sử mãi mãi tại sông Bạch Đằng...",
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
				{
					ID:       "q_nq1_3",
					Type:     "multiple_choice",
					Question: "Ngô Quyền giành chiến thắng vào năm nào?",
					Options:  []string{"936", "938", "940", "942"},
					Answer:   "938",
					Points:   15,
				},
				{
					ID:       "q_nq1_4",
					Type:     "fill_blank",
					Question: "Chiến thắng của Ngô Quyền đã kết thúc gần một _____ năm Bắc thuộc.",
					Answer:   "thiên niên kỷ",
					Points:   20,
				},
				{
					ID:       "q_nq1_5",
					Type:     "multiple_choice",
					Question: "Ngô Quyền có chức vụ gì trước khi trở thành vua?",
					Options:  []string{"Thái sư", "Tướng lĩnh", "Quan tri huyện", "Thái úy"},
					Answer:   "Tướng lĩnh",
					Points:   15,
				},
				// {
				// 	ID:       "q_nq1_6",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự thời gian các giai đoạn lịch sử:",
				// 	Options:  []string{"Bắc thuộc lần 1", "Các cuộc khởi nghĩa", "Chiến thắng Bạch Đằng", "Độc lập tự chủ"},
				// 	Answer:   []string{"Bắc thuộc lần 1", "Các cuộc khởi nghĩa", "Chiến thắng Bạch Đằng", "Độc lập tự chủ"},
				// 	Points:   25,
				// },
			}),
			XPReward:  80,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Lý Thái Tổ lessons
		{
			ID:          "lesson_ly_thai_to_1",
			CharacterID: "char_ly_thai_to",
			Title:       "Việc dời đô về Thăng Long",
			Order:       1,
			Story:       "Năm 1010, Hoàng đế Lý Thái Tổ đã đưa ra một quyết định trọng đại - di chuyển kinh đô từ Hoa Lư đến vị trí của Hà Nội ngày nay. Ngài đặt tên là Thăng Long, có nghĩa là 'Rồng Bay Lên', sau khi chứng kiến một con rồng vàng bay lên từ sông Hồng...",
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
				{
					ID:       "q_ltt1_3",
					Type:     "fill_blank",
					Question: "Trước khi dời về Thăng Long, kinh đô cũ là _____.",
					Answer:   "Hoa Lư",
					Points:   15,
				},
				{
					ID:       "q_ltt1_4",
					Type:     "multiple_choice",
					Question: "Theo truyền thuyết, Lý Thái Tổ thấy điều gì khi đến Thăng Long?",
					Options:  []string{"Phượng hoàng vàng", "Rồng vàng bay lên", "Kỳ lân xuất hiện", "Rùa thần nổi lên"},
					Answer:   "Rồng vàng bay lên",
					Points:   15,
				},
				{
					ID:       "q_ltt1_5",
					Type:     "multiple_choice",
					Question: "Thăng Long ngày nay là thành phố nào?",
					Options:  []string{"Hồ Chí Minh", "Hà Nội", "Đà Nẵng", "Huế"},
					Answer:   "Hà Nội",
					Points:   10,
				},
				{
					ID:       "q_ltt1_6",
					Type:     "multiple_choice",
					Question: "Lý Thái Tổ là hoàng đế đầu tiên của triều đại nào?",
					Options:  []string{"Triều đại Trần", "Triều đại Lý", "Triều đại Lê", "Triều đại Ngô"},
					Answer:   "Triều đại Lý",
					Points:   15,
				},
				{
					ID:       "q_ltt1_7",
					Type:     "connect",
					Question: "Nối các yếu tố với ý nghĩa trong việc dời đô:",
					Options:  []string{"Hoa Lư", "Thăng Long", "Rồng vàng", "Sông Hồng"},
					Answer:   map[string]string{"Hoa Lư": "Kinh đô cũ", "Thăng Long": "Kinh đô mới", "Rồng vàng": "Điềm lành", "Sông Hồng": "Vị trí địa lý thuận lợi"},
					Points:   25,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Hoa Lư": "Kinh đô cũ", "Thăng Long": "Kinh đô mới", "Rồng vàng": "Điềm lành", "Sông Hồng": "Vị trí địa lý thuận lợi"}},
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
			ID:          "lesson_ho_chi_minh_1",
			CharacterID: "char_ho_chi_minh",
			Title:       "Người thanh niên yêu nước",
			Order:       1,
			Story:       "Sinh ra với tên Nguyễn Sinh Cung vào năm 1890, vị lãnh tụ tương lai của Việt Nam lớn lên trong việc chứng kiến sự áp bức thuộc địa của Pháp. Khi còn trẻ, ông rời Việt Nam trên một tàu hơi nước của Pháp, bắt đầu cuộc hành trình đưa ông đi khắp thế giới và hình thành lý tưởng cách mạng...",
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
				{
					ID:       "q_hcm1_3",
					Type:     "multiple_choice",
					Question: "Hồ Chí Minh rời Việt Nam bằng phương tiện gì?",
					Options:  []string{"Máy bay", "Tàu hơi nước của Pháp", "Tàu thuyền", "Đường bộ"},
					Answer:   "Tàu hơi nước của Pháp",
					Points:   15,
				},
				{
					ID:       "q_hcm1_4",
					Type:     "multiple_choice",
					Question: "Việt Nam bị ai áp bức thuộc địa khi Hồ Chí Minh còn trẻ?",
					Options:  []string{"Nhật Bản", "Pháp", "Trung Quốc", "Mỹ"},
					Answer:   "Pháp",
					Points:   10,
				},
				{
					ID:       "q_hcm1_5",
					Type:     "multiple_choice",
					Question: "Tên khác của Hồ Chí Minh trong thời gian hoạt động cách mạng là gì?",
					Options:  []string{"Nguyễn Tất Thành", "Nguyễn Ái Quốc", "Phan Bội Châu", "Nguyễn Du"},
					Answer:   "Nguyễn Ái Quốc",
					Points:   15,
				},
				// {
				// 	ID:       "q_hcm1_6",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự thời gian cuộc đời trẻ của Hồ Chí Minh:",
				// 	Options:  []string{"Sinh ra với tên Nguyễn Sinh Cung", "Chứng kiến áp bức thuộc địa", "Rời Việt Nam trên tàu", "Hình thành lý tưởng cách mạng"},
				// 	Answer:   []string{"Sinh ra với tên Nguyễn Sinh Cung", "Chứng kiến áp bức thuộc địa", "Rời Việt Nam trên tàu", "Hình thành lý tưởng cách mạng"},
				// 	Points:   20,
				// },
				{
					ID:       "q_hcm1_7",
					Type:     "connect",
					Question: "Nối các giai đoạn với đặc điểm của Hồ Chí Minh:",
					Options:  []string{"Thời thơ ấu", "Thời thanh niên", "Du học", "Hoạt động cách mạng"},
					Answer:   map[string]string{"Thời thơ ấu": "Tên Nguyễn Sinh Cung", "Thời thanh niên": "Chứng kiến áp bức", "Du học": "Đi khắp thế giới", "Hoạt động cách mạng": "Tên Nguyễn Ái Quốc"},
					Points:   25,
					Metadata: map[string]interface{}{"pairs": map[string]string{"Thời thơ ấu": "Tên Nguyễn Sinh Cung", "Thời thanh niên": "Chứng kiến áp bức", "Du học": "Đi khắp thế giới", "Hoạt động cách mạng": "Tên Nguyễn Ái Quốc"}},
				},
			}),
			XPReward:  85,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "lesson_ho_chi_minh_2",
			CharacterID: "char_ho_chi_minh",
			Title:       "Tuyên ngôn Độc lập",
			Order:       2,
			Story:       "Ngày 2 tháng 9 năm 1945, tại Quảng trường Ba Đình, Hồ Chí Minh đã đọc Tuyên ngôn Độc lập, thành lập nước Việt Nam Dân chủ Cộng hòa. Lời mở đầu của ông trích dẫn Tuyên ngôn Độc lập của Mỹ: 'Tất cả mọi người sinh ra đều bình đẳng...'",
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
				{
					ID:       "q_hcm2_3",
					Type:     "fill_blank",
					Question: "Tuyên ngôn Độc lập thành lập nước _____ _____ _____ _____.",
					Answer:   "Việt Nam Dân chủ Cộng hòa",
					Points:   20,
				},
				{
					ID:       "q_hcm2_4",
					Type:     "multiple_choice",
					Question: "Hồ Chí Minh trích dẫn Tuyên ngôn Độc lập của nước nào?",
					Options:  []string{"Pháp", "Anh", "Mỹ", "Nga"},
					Answer:   "Mỹ",
					Points:   15,
				},
				{
					ID:       "q_hcm2_5",
					Type:     "fill_blank",
					Question: "Câu nổi tiếng được trích dẫn: 'Tất cả mọi người sinh ra đều _____'.",
					Answer:   "bình đẳng",
					Points:   15,
				},
				{
					ID:       "q_hcm2_6",
					Type:     "multiple_choice",
					Question: "Ngày 2/9 hiện tại được gọi là gì?",
					Options:  []string{"Ngày Giải phóng", "Ngày Quốc khánh", "Ngày Thống nhất", "Ngày Cách mạng"},
					Answer:   "Ngày Quốc khánh",
					Points:   10,
				},
				// {
				// 	ID:       "q_hcm2_7",
				// 	Type:     "drag_drop",
				// 	Question: "Sắp xếp theo thứ tự các sự kiện dẫn đến Tuyên ngôn Độc lập:",
				// 	Options:  []string{"Kết thúc Thế chiến II", "Cách mạng tháng Tám", "Tuyên ngôn Độc lập", "Thành lập nước VNDCCH"},
				// 	Answer:   []string{"Kết thúc Thế chiến II", "Cách mạng tháng Tám", "Tuyên ngôn Độc lập", "Thành lập nước VNDCCH"},
				// 	Points:   25,
				// },
				{
					ID:       "q_hcm2_8",
					Type:     "connect",
					Question: "Nối các yếu tố với ý nghĩa trong Tuyên ngôn Độc lập:",
					Options:  []string{"2/9/1945", "Ba Đình", "VNDCCH", "Bình đẳng"},
					Answer:   map[string]string{"2/9/1945": "Ngày tuyên bố", "Ba Đình": "Địa điểm lịch sử", "VNDCCH": "Tên nước mới", "Bình đẳng": "Giá trị nhân văn"},
					Points:   30,
					Metadata: map[string]interface{}{"pairs": map[string]string{"2/9/1945": "Ngày tuyên bố", "Ba Đình": "Địa điểm lịch sử", "VNDCCH": "Tên nước mới", "Bình đẳng": "Giá trị nhân văn"}},
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

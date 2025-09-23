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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(-2879),
			DeathYear:   intPtr(-258),
			Description: "Vị vua huyền thoại sáng lập nhà nước Việt Nam đầu tiên - Văn Lang. Được biết đến là vua đầu tiên của Việt Nam, người đã xây dựng nền tảng cho nền văn minh Việt Nam ở vùng đồng bằng sông Hồng.",
			FamousQuote: "Con rồng cháu tiên",
			Achievements: jsonArray([]string{
				"Sáng lập vương quốc Văn Lang",
				"Thiết lập triều đại Hùng Vương",
				"Tạo nền tảng cho văn hóa Việt Nam",
				"Thống nhất các bộ lạc Lạc Việt",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(974),
			DeathYear:   intPtr(1028),
			Description: "Người sáng lập triều đại Lý và xây dựng Thăng Long (Hà Nội ngày nay). Ngài đã di dời kinh đô từ Hoa Lư về Thăng Long, mở ra thời kỳ hoàng kim của văn hóa Việt Nam và Phật giáo.",
			FamousQuote: "Thăng Long hữu đức khí",
			Achievements: jsonArray([]string{
				"Sáng lập triều đại Lý năm 1009",
				"Thiết lập Thăng Long làm kinh đô",
				"Thúc đẩy Phật giáo và giáo dục",
				"Thống nhất và củng cố Việt Nam",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(1228),
			DeathYear:   intPtr(1300),
			Description: "Nhà quân sự thiên tài vĩ đại nhất trong lịch sử Việt Nam. Đã thành công bảo vệ Việt Nam trước ba cuộc xâm lược của quân Mông Cổ, sử dụng chiến thuật du kích và kiến thức ưu việt về địa hình địa phương.",
			FamousQuote: "Thà chết vì tổ quốc chứ không sống làm nô lệ",
			Achievements: jsonArray([]string{
				"Đánh bại ba cuộc xâm lược của quân Mông Cổ (1258, 1285, 1287-1288)",
				"Phát triển chiến thuật du kích đổi mới",
				"Bảo vệ nền độc lập của Việt Nam",
				"Viết các luận văn quân sự về chiến lược",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(1385),
			DeathYear:   intPtr(1433),
			Description: "Anh hùng giải phóng Việt Nam khỏi sự chiếm đóng của nhà Minh Trung Quốc. Sáng lập triều đại Lê sau mười năm kháng chiến, trở thành hoàng đế Lê Thái Tổ.",
			FamousQuote: "Nam quốc sơn hà Nam đế cư",
			Achievements: jsonArray([]string{
				"Lãnh đạo cuộc khởi nghĩa thành công chống lại nhà Minh Trung Quốc (1418-1428)",
				"Sáng lập triều đại Hậu Lê",
				"Giải phóng Việt Nam khỏi sự cai trị ngoại bang",
				"Thiết lập Bình Ngô Đại Cáo",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(1380),
			DeathYear:   intPtr(1442),
			Description: "Nhà chiến lược, thi sĩ và ngoại giao tài ba phục vụ Lê Lợi. Nổi tiếng với các tác phẩm văn học và bản tuyên ngôn nổi tiếng 'Bình Ngô Đại Cáo'.",
			FamousQuote: "Dĩ nhân thắng bạo, dĩ nghĩa thắng phi nghĩa",
			Achievements: jsonArray([]string{
				"Tác giả Bình Ngô Đại Cáo (Tuyên ngôn Vĩ đại về Chiến thắng)",
				"Cố vấn trưởng cho Lê Lợi trong cuộc chiến độc lập",
				"Thi sĩ và nhà văn học nổi tiếng",
				"Tiên phong trong việc viết ngoại giao Việt Nam",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(1753),
			DeathYear:   intPtr(1792),
			Description: "Hoàng đế Quang Trung, lãnh tụ của cuộc khởi nghĩa Tây Sơn. Đánh bại cuộc xâm lược của nhà Thanh và thực hiện các cải cách quan trọng bao gồm thúc đẩy ngôn ngữ và văn hóa Việt Nam.",
			FamousQuote: "Bắc định Thanh quân, nam bình Chúa Nguyễn",
			Achievements: jsonArray([]string{
				"Lãnh đạo cuộc khởi nghĩa Tây Sơn",
				"Đánh bại quân Thanh tại Đống Đa (1789)",
				"Thúc đẩy tiếng Việt trong giáo dục",
				"Thực hiện các cải cách xã hội tiến bộ",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(12),
			DeathYear:   intPtr(43),
			Description: "Hai Bà Trưng - Trưng Trắc và Trưng Nhị - lãnh đạo cuộc nổi dậy lớn đầu tiên chống lại sự thống trị của Trung Quốc. Họ đã thiết lập một vương quốc độc lập trong ba năm.",
			FamousQuote: "Trước để trừ giặc, sau để cảnh báo hậu thế",
			Achievements: jsonArray([]string{
				"Lãnh đạo cuộc nổi dậy thành công chống lại triều đại Hán Trung Quốc",
				"Thiết lập vương quốc Việt Nam độc lập (40-43 SCN)",
				"Biểu tượng của sức mạnh phụ nữ Việt Nam",
				"Những người cầm quyền Việt Nam đầu tiên được ghi nhận sau sự chiếm đóng của Trung Quốc",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(1019),
			DeathYear:   intPtr(1105),
			Description: "Đại tướng vĩ đại của triều Lý, người đã thành công bảo vệ Việt Nam trước cuộc xâm lược của nhà Tống Trung Quốc. Nổi tiếng với những đổi mới quân sự và bài thơ 'Nam quốc sơn hà'.",
			FamousQuote: "Nam quốc sơn hà Nam đế cư, Tiệt nhiên định phận tại thiên thư",
			Achievements: jsonArray([]string{
				"Đánh bại cuộc xâm lược của nhà Tống Trung Quốc",
				"Tác giả bài thơ yêu nước nổi tiếng 'Nam quốc sơn hà'",
				"Làm nhiếp chính trong thời niên thiếu của Lý Nhân Tông",
				"Củng cố biên giới Việt Nam",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(897),
			DeathYear:   intPtr(944),
			Description: "Người sáng lập triều đại Ngô, chấm dứt gần 1000 năm thống trị của Trung Quốc. Giành chiến thắng quyết định tại trận Bạch Đằng chống lại hạm đội Nam Hán.",
			FamousQuote: "Đại Việt độc lập",
			Achievements: jsonArray([]string{
				"Thắng trận Bạch Đằng (938)",
				"Chấm dứt sự thống trị của Trung Quốc sau 1000 năm",
				"Sáng lập triều đại Ngô",
				"Thiết lập nền độc lập Việt Nam",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(924),
			DeathYear:   intPtr(979),
			Description: "Hoàng đế đầu tiên của Việt Nam thống nhất, được biết đến với tên Đinh Tiên Hoàng. Thành lập vương quốc Đại Cồ Việt và mang lại sự ổn định sau nhiều năm hỗn loạn.",
			FamousQuote: "Đại Cồ Việt Hoàng Đế",
			Achievements: jsonArray([]string{
				"Thống nhất Việt Nam dưới vương quốc Đại Cồ Việt",
				"Hoàng đế đầu tiên của Việt Nam độc lập",
				"Thiết lập kinh đô tại Hoa Lư",
				"Tạo cấu trúc chính phủ có tổ chức",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(1442),
			DeathYear:   intPtr(1497),
			Description: "Hoàng đế vĩ đại nhất của triều Lê, nổi tiếng với việc mở rộng lãnh thổ và cải cách pháp luật. Tạo ra bộ luật Hồng Đức và thúc đẩy văn học và nghệ thuật.",
			FamousQuote: "Minh đức tân dân",
			Achievements: jsonArray([]string{
				"Tạo ra bộ luật Hồng Đức toàn diện",
				"Mở rộng lãnh thổ đến vương quốc Champa",
				"Thúc đẩy giáo dục Nho giáo và công vụ",
				"Thời kỳ hoàng kim của văn học và nghệ thuật Việt Nam",
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
			Rarity:      "Thường",
			BirthYear:   intPtr(1336),
			DeathYear:   intPtr(1407),
			Description: "Nhà cải cách gây tranh cãi, người sáng lập triều đại Hồ ngắn ngủi. Nổi tiếng với các cải cách tiến bộ nhưng cũng vì những sự kiện dẫn đến cuộc xâm lược của nhà Minh Trung Quốc.",
			FamousQuote: "Cải cách duy tân",
			Achievements: jsonArray([]string{
				"Thực hiện cải cách phân phối đất đai",
				"Thúc đẩy hệ thống tiền giấy",
				"Kỹ thuật nông nghiệp tiên tiến",
				"Sáng lập triều đại Hồ (1400-1407)",
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
			Rarity:      "Thường",
			BirthYear:   intPtr(1483),
			DeathYear:   intPtr(1541),
			Description: "Người sáng lập triều đại Mạc, đã chiếm quyền lực từ triều Lê. Nổi tiếng với kỹ năng quân sự nhưng tính chính thống của ông thường bị tranh cãi.",
			FamousQuote: "Thiên mệnh tại ta",
			Achievements: jsonArray([]string{
				"Sáng lập triều đại Mạc",
				"Chỉ huy quân sự tài ba",
				"Kiểm soát miền Bắc Việt Nam",
				"Thiết lập quan hệ ngoại giao với Trung Quốc",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(1762),
			DeathYear:   intPtr(1820),
			Description: "Sáng lập triều đại Nguyễn và thống nhất Việt Nam dưới danh hiệu hoàng đế Gia Long. Thiết lập Huế làm kinh đô và tạo ra biên giới hiện đại của Việt Nam.",
			FamousQuote: "Thống nhất giang sơn",
			Achievements: jsonArray([]string{
				"Thống nhất Việt Nam từ Bắc chí Nam",
				"Sáng lập triều đại Nguyễn (1802-1945)",
				"Thiết lập Huế làm kinh đô hoàng gia",
				"Tạo ra sự thống nhất lãnh thổ Việt Nam hiện đại",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(1867),
			DeathYear:   intPtr(1940),
			Description: "Nhà hoạt động dân tộc chủ nghĩa và đấu tranh độc lập tiên phong. Lãnh đạo các phong trào kháng chiến ban đầu chống lại sự cai trị thuộc địa của Pháp và thúc đẩy giáo dục và cải cách hiện đại.",
			FamousQuote: "Việt Nam vong quốc sử",
			Achievements: jsonArray([]string{
				"Sáng lập Việt Nam Duy Tân Hội",
				"Lãnh đạo phong trào Đông Du sang Nhật Bản",
				"Viết các tác phẩm văn học dân tộc chủ nghĩa có ảnh hưởng",
				"Tiên phong của phong trào độc lập Việt Nam",
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
			Rarity:      "Thường",
			BirthYear:   intPtr(1872),
			DeathYear:   intPtr(1926),
			Description: "Học giả cải cách ủng hộ hiện đại hóa và giáo dục. Thúc đẩy kháng chiến hòa bình và cải cách văn hóa trong thời kỳ thuộc địa Pháp.",
			FamousQuote: "Dân trí là gốc của mọi việc",
			Achievements: jsonArray([]string{
				"Ủng hộ cải cách giáo dục",
				"Thúc đẩy các phương pháp kháng chiến hòa bình",
				"Sáng lập báo chí hiện đại Việt Nam",
				"Ảnh hưởng đến phong trào thức tỉnh trí thức",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(1765),
			DeathYear:   intPtr(1820),
			Description: "Thi sĩ vĩ đại nhất của Việt Nam, tác giả của thiên trường ca 'Truyện Kiều'. Tác phẩm của ông đại diện cho đỉnh cao của văn học cổ điển Việt Nam.",
			FamousQuote: "Trăm năm trong cõi người ta, chữ tài chữ mệnh khéo là ghét nhau",
			Achievements: jsonArray([]string{
				"Tác giả Truyện Kiều, kiệt tác của văn học Việt Nam",
				"Bậc thầy thơ chữ Nôm",
				"Ảnh hưởng đến bản sắc văn hóa Việt Nam",
				"UNESCO công nhận đóng góp cho văn học thế giới",
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
			Rarity:      "Hiếm",
			BirthYear:   intPtr(225),
			DeathYear:   intPtr(248),
			Description: "Nữ chiến binh lãnh đạo cuộc nổi dậy chống lại sự chiếm đóng của nhà Ngô Trung Quốc vào thế kỷ thứ 3. Nổi tiếng với lòng dũng cảm và quyết tâm chống lại sự thống trị ngoại bang.",
			FamousQuote: "Tôi muốn cưỡi cơn gió mạnh, đạp sóng dữ, chém cá kình ở biển Đông",
			Achievements: jsonArray([]string{
				"Lãnh đạo cuộc nổi dậy chống lại triều đại Ngô Trung Quốc",
				"Biểu tượng của chủ nghĩa anh hùng phụ nữ Việt Nam",
				"Đấu tranh cho độc lập Việt Nam trong thế kỷ thứ 3",
				"Truyền cảm hứng cho các thế hệ yêu nước sau này",
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
			Rarity:      "Thường",
			BirthYear:   intPtr(1933),
			DeathYear:   intPtr(1952),
			Description: "Thanh niên cách mạng trở thành biểu tượng của sự kháng chiến trong thời kỳ thuộc địa Pháp. Bị hành quyết ở tuổi 19 vì các hoạt động chống thực dân.",
			FamousQuote: "Tôi chết vì Tổ quốc, lòng tôi vui sướng lắm",
			Achievements: jsonArray([]string{
				"Thành viên tích cực của phong trào kháng chiến",
				"Biểu tượng của lòng yêu nước thanh niên Việt Nam",
				"Hy sinh mạng sống vì sự nghiệp độc lập",
				"Truyền cảm hứng cho kháng chiến chống thực dân",
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
			Rarity:      "Huyền thoại",
			BirthYear:   intPtr(1890),
			DeathYear:   intPtr(1969),
			Description: "Cha đẻ sáng lập nước Việt Nam hiện đại và lãnh tụ của phong trào độc lập. Lãnh đạo cuộc đấu tranh của đất nước chống lại chủ nghĩa thực dân Pháp và sự can thiệp của Mỹ.",
			FamousQuote: "Không có gì quý hơn độc lập tự do",
			Achievements: jsonArray([]string{
				"Sáng lập nước Việt Nam Dân chủ Cộng hòa",
				"Lãnh đạo phong trào độc lập chống Pháp",
				"Tuyên bố độc lập của Việt Nam (1945)",
				"Cha đẻ của dân tộc Việt Nam hiện đại",
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

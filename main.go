package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// İl yapısı
type Il struct {
	ID    string `json:"id"`
	Adi   string `json:"adi"`
	AdiEn string `json:"adiEn"`
}

// İlçe yapısı
type Ilce struct {
	IlceUrl   string `json:"IlceUrl"`
	IlceAdi   string `json:"IlceAdi"`
	IlceAdiEn string `json:"IlceAdiEn"`
	IlceID    string `json:"IlceID"`
}

// Türkiye veri yapısı
type TurkeyData struct {
	Ulke struct {
		ID    string `json:"id"`
		Adi   string `json:"adi"`
		AdiEn string `json:"adiEn"`
	} `json:"ulke"`
	Iller []Il `json:"iller"`
}

// API yanıt yapısı
type StateRegionResponse struct {
	Result          interface{} `json:"Result"`
	CountryList     interface{} `json:"CountryList"`
	StateList       interface{} `json:"StateList"`
	StateRegionList []Ilce      `json:"StateRegionList"`
	HasStateList    bool        `json:"HasStateList"`
}

// Namaz vakti yapısı
type NamazVakti struct {
	VakitAdi string `json:"VakitAdi"`
	Vakit    string `json:"Vakit"`
}

// Günlük vakitler yapısı
type GunlukVakitler struct {
	Tarih      string       `json:"tarih"`
	TarihISO   string       `json:"tarih_iso"`
	HicriTarih string       `json:"hicriTarih"`
	Vakitler   []NamazVakti `json:"vakitler"`
}

// Vakit tipi enum
type VakitTipi string

const (
	Gunluk   VakitTipi = "gunluk"
	Haftalik VakitTipi = "haftalik"
	Yillik   VakitTipi = "yillik"
)

// Sonuç yapısı
type Sonuc struct {
	Il               string                 `json:"il"`
	IlID             string                 `json:"il_id"`
	Ilce             string                 `json:"ilce"`
	IlceID           string                 `json:"ilce_id"`
	VakitTipi        VakitTipi              `json:"vakit_tipi"`
	Tarih            string                 `json:"tarih"`
	GunlukVakitler   []NamazVakti           `json:"gunluk_vakitler,omitempty"`
	HaftalikVakitler []GunlukVakitler       `json:"haftalik_vakitler,omitempty"`
	YillikVakitler   []GunlukVakitler       `json:"yillik_vakitler,omitempty"`
	Istatistikler    map[string]interface{} `json:"istatistikler,omitempty"`
}

// İlçe listesini al
func getIlceListesi(stateID string) ([]Ilce, error) {
	url := fmt.Sprintf("https://namazvakitleri.diyanet.gov.tr/tr-TR/home/GetRegList?ChangeType=state&CountryId=2&Culture=tr-TR&StateId=%s", stateID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği hatası: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP hatası! durum: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("yanıt okuma hatası: %v", err)
	}

	// Debug için yanıtı yazdır (sadece geliştirme sırasında)
	// fmt.Printf("API Yanıtı: %s\n", string(body))

	var data StateRegionResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("JSON parse hatası: %v", err)
	}

	return data.StateRegionList, nil
}

// Türkiye verilerini JSON dosyasından yükle
// Embedded Turkey data
var embeddedTurkeyData = TurkeyData{
	Ulke: struct {
		ID    string `json:"id"`
		Adi   string `json:"adi"`
		AdiEn string `json:"adiEn"`
	}{
		ID:    "2",
		Adi:   "Türkiye",
		AdiEn: "Turkey",
	},
	Iller: []Il{
		{ID: "500", Adi: "ADANA", AdiEn: "ADANA"},
		{ID: "501", Adi: "ADIYAMAN", AdiEn: "ADIYAMAN"},
		{ID: "502", Adi: "AFYONKARAHİSAR", AdiEn: "AFYONKARAHISAR"},
		{ID: "503", Adi: "AĞRI", AdiEn: "AGRI"},
		{ID: "504", Adi: "AKSARAY", AdiEn: "AKSARAY"},
		{ID: "505", Adi: "AMASYA", AdiEn: "AMASYA"},
		{ID: "506", Adi: "ANKARA", AdiEn: "ANKARA"},
		{ID: "507", Adi: "ANTALYA", AdiEn: "ANTALYA"},
		{ID: "508", Adi: "ARDAHAN", AdiEn: "ARDAHAN"},
		{ID: "509", Adi: "ARTVİN", AdiEn: "ARTVIN"},
		{ID: "510", Adi: "AYDIN", AdiEn: "AYDIN"},
		{ID: "511", Adi: "BALIKESİR", AdiEn: "BALIKESIR"},
		{ID: "512", Adi: "BARTIN", AdiEn: "BARTIN"},
		{ID: "513", Adi: "BATMAN", AdiEn: "BATMAN"},
		{ID: "514", Adi: "BAYBURT", AdiEn: "BAYBURT"},
		{ID: "515", Adi: "BİLECİK", AdiEn: "BILECIK"},
		{ID: "516", Adi: "BİNGÖL", AdiEn: "BINGOL"},
		{ID: "517", Adi: "BİTLİS", AdiEn: "BITLIS"},
		{ID: "518", Adi: "BOLU", AdiEn: "BOLU"},
		{ID: "519", Adi: "BURDUR", AdiEn: "BURDUR"},
		{ID: "520", Adi: "BURSA", AdiEn: "BURSA"},
		{ID: "521", Adi: "ÇANAKKALE", AdiEn: "CANAKKALE"},
		{ID: "522", Adi: "ÇANKIRI", AdiEn: "CANKIRI"},
		{ID: "523", Adi: "ÇORUM", AdiEn: "CORUM"},
		{ID: "524", Adi: "DENİZLİ", AdiEn: "DENIZLI"},
		{ID: "525", Adi: "DİYARBAKIR", AdiEn: "DIYARBAKIR"},
		{ID: "526", Adi: "DÜZCE", AdiEn: "DUZCE"},
		{ID: "527", Adi: "EDİRNE", AdiEn: "EDIRNE"},
		{ID: "528", Adi: "ELAZIĞ", AdiEn: "ELAZIG"},
		{ID: "529", Adi: "ERZİNCAN", AdiEn: "ERZINCAN"},
		{ID: "530", Adi: "ERZURUM", AdiEn: "ERZURUM"},
		{ID: "531", Adi: "ESKİŞEHİR", AdiEn: "ESKISEHIR"},
		{ID: "532", Adi: "GAZİANTEP", AdiEn: "GAZIANTEP"},
		{ID: "533", Adi: "GİRESUN", AdiEn: "GIRESUN"},
		{ID: "534", Adi: "GÜMÜŞHANE", AdiEn: "GUMUSHANE"},
		{ID: "535", Adi: "HAKKARİ", AdiEn: "HAKKARI"},
		{ID: "536", Adi: "HATAY", AdiEn: "HATAY"},
		{ID: "537", Adi: "IĞDIR", AdiEn: "IGDIR"},
		{ID: "538", Adi: "ISPARTA", AdiEn: "ISPARTA"},
		{ID: "539", Adi: "İSTANBUL", AdiEn: "ISTANBUL"},
		{ID: "540", Adi: "İZMİR", AdiEn: "IZMIR"},
		{ID: "541", Adi: "KAHRAMANMARAŞ", AdiEn: "KAHRAMANMARAS"},
		{ID: "542", Adi: "KARABÜK", AdiEn: "KARABUK"},
		{ID: "543", Adi: "KARAMAN", AdiEn: "KARAMAN"},
		{ID: "544", Adi: "KARS", AdiEn: "KARS"},
		{ID: "545", Adi: "KASTAMONU", AdiEn: "KASTAMONU"},
		{ID: "546", Adi: "KAYSERİ", AdiEn: "KAYSERI"},
		{ID: "547", Adi: "KİLİS", AdiEn: "KILIS"},
		{ID: "548", Adi: "KIRIKKALE", AdiEn: "KIRIKKALE"},
		{ID: "549", Adi: "KIRKLARELİ", AdiEn: "KIRKLARELI"},
		{ID: "550", Adi: "KIRŞEHİR", AdiEn: "KIRSEHIR"},
		{ID: "551", Adi: "KOCAELİ", AdiEn: "KOCAELI"},
		{ID: "552", Adi: "KONYA", AdiEn: "KONYA"},
		{ID: "553", Adi: "KÜTAHYA", AdiEn: "KUTAHYA"},
		{ID: "554", Adi: "MALATYA", AdiEn: "MALATYA"},
		{ID: "555", Adi: "MANİSA", AdiEn: "MANISA"},
		{ID: "556", Adi: "MARDİN", AdiEn: "MARDIN"},
		{ID: "557", Adi: "MERSİN", AdiEn: "MERSIN"},
		{ID: "558", Adi: "MUĞLA", AdiEn: "MUGLA"},
		{ID: "559", Adi: "MUŞ", AdiEn: "MUS"},
		{ID: "560", Adi: "NEVŞEHİR", AdiEn: "NEVSEHIR"},
		{ID: "561", Adi: "NİĞDE", AdiEn: "NIGDE"},
		{ID: "562", Adi: "ORDU", AdiEn: "ORDU"},
		{ID: "563", Adi: "OSMANİYE", AdiEn: "OSMANIYE"},
		{ID: "564", Adi: "RİZE", AdiEn: "RIZE"},
		{ID: "565", Adi: "SAKARYA", AdiEn: "SAKARYA"},
		{ID: "566", Adi: "SAMSUN", AdiEn: "SAMSUN"},
		{ID: "567", Adi: "ŞANLIURFA", AdiEn: "SANLIURFA"},
		{ID: "568", Adi: "SİİRT", AdiEn: "SIIRT"},
		{ID: "569", Adi: "SİNOP", AdiEn: "SINOP"},
		{ID: "570", Adi: "ŞIRNAK", AdiEn: "SIRNAK"},
		{ID: "571", Adi: "SİVAS", AdiEn: "SIVAS"},
		{ID: "572", Adi: "TEKİRDAĞ", AdiEn: "TEKIRDAG"},
		{ID: "573", Adi: "TOKAT", AdiEn: "TOKAT"},
		{ID: "574", Adi: "TRABZON", AdiEn: "TRABZON"},
		{ID: "575", Adi: "TUNCELİ", AdiEn: "TUNCELI"},
		{ID: "576", Adi: "UŞAK", AdiEn: "USAK"},
		{ID: "577", Adi: "VAN", AdiEn: "VAN"},
		{ID: "578", Adi: "YALOVA", AdiEn: "YALOVA"},
		{ID: "579", Adi: "YOZGAT", AdiEn: "YOZGAT"},
		{ID: "580", Adi: "ZONGULDAK", AdiEn: "ZONGULDAK"},
	},
}

func loadTurkeyData() (*TurkeyData, error) {
	return &embeddedTurkeyData, nil
}

// İl adından ID bul
func findIlIDByName(turkeyData *TurkeyData, ilAdi string) (string, error) {
	for _, il := range turkeyData.Iller {
		if strings.EqualFold(il.Adi, ilAdi) {
			return il.ID, nil
		}
	}
	return "", fmt.Errorf("il bulunamadı: %s", ilAdi)
}

// İl ID'sinden ad bul
func findIlNameByID(turkeyData *TurkeyData, ilID string) (string, error) {
	for _, il := range turkeyData.Iller {
		if il.ID == ilID {
			return il.Adi, nil
		}
	}
	return "", fmt.Errorf("il bulunamadı: %s", ilID)
}

// Dosya yolu oluştur
func createFilePath(ilAdi, ilceAdi, vakitTipi string) string {
	// Türkçe karakterleri ve boşlukları temizle
	ilAdi = cleanFileName(ilAdi)
	ilceAdi = cleanFileName(ilceAdi)

	return fmt.Sprintf("vakitler/%s/%s/%s.json", ilAdi, ilceAdi, vakitTipi)
}

// Dosya adı temizleme
func cleanFileName(name string) string {
	// Türkçe karakterleri değiştir
	replacements := map[string]string{
		"İ": "I", "ı": "i", "Ğ": "G", "ğ": "g", "Ü": "U", "ü": "u",
		"Ş": "S", "ş": "s", "Ö": "O", "ö": "o", "Ç": "C", "ç": "c",
		" ": "_", "-": "_", "&": "ve",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Sadece harf, rakam ve alt çizgi bırak
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	result = reg.ReplaceAllString(result, "")

	return strings.ToLower(result)
}

// Klasör oluştur
func createDirectory(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("klasör oluşturulamadı: %v", err)
	}
	return nil
}

// Türkçe tarih formatını ISO formatına çevir
func parseTarihToISO(tarih string) string {
	// Türkçe ay isimleri
	aylar := map[string]string{
		"Ocak": "01", "Şubat": "02", "Mart": "03", "Nisan": "04",
		"Mayıs": "05", "Haziran": "06", "Temmuz": "07", "Ağustos": "08",
		"Eylül": "09", "Ekim": "10", "Kasım": "11", "Aralık": "12",
	}

	// Tarih formatı: "03 Ağustos 2025 Pazar"
	parts := strings.Fields(tarih)
	if len(parts) >= 3 {
		gun := parts[0]
		ay := parts[1]
		yil := parts[2]

		// Ay numarasını bul
		if ayNum, ok := aylar[ay]; ok {
			// Gün numarasını 2 haneli yap
			if len(gun) == 1 {
				gun = "0" + gun
			}
			return fmt.Sprintf("%s-%s-%s", yil, ayNum, gun)
		}
	}

	// Parse edilemezse boş string döndür
	return ""
}

// Namaz vakitleri HTML'ini al
func getNamazVakitleriHTML(ilceID string) (string, error) {
	url := fmt.Sprintf("https://namazvakitleri.diyanet.gov.tr/tr-TR/%s", ilceID)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP isteği hatası: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP hatası! durum: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("yanıt okuma hatası: %v", err)
	}

	return string(body), nil
}

// Günlük vakitleri parse et
func parseGunlukVakitler(html string) []NamazVakti {
	vakitler := []NamazVakti{}

	// JavaScript değişkenlerinden vakitleri al
	patterns := map[string]string{
		"İmsak":  `_imsakTime\s*=\s*"(\d{2}:\d{2})"`,
		"Güneş":  `_gunesTime\s*=\s*"(\d{2}:\d{2})"`,
		"Öğle":   `_ogleTime\s*=\s*"(\d{2}:\d{2})"`,
		"İkindi": `_ikindiTime\s*=\s*"(\d{2}:\d{2})"`,
		"Akşam":  `_aksamTime\s*=\s*"(\d{2}:\d{2})"`,
		"Yatsı":  `_yatsiTime\s*=\s*"(\d{2}:\d{2})"`,
	}

	for vakitAdi, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			vakitler = append(vakitler, NamazVakti{
				VakitAdi: vakitAdi,
				Vakit:    matches[1],
			})
		}
	}

	return vakitler
}

// Haftalık vakitleri parse et
func parseNamazVakitleri(html string) []GunlukVakitler {
	vakitler := []GunlukVakitler{}

	// HTML tablo satırlarını bul
	tableRowRegex := regexp.MustCompile(`<tr>\s*<td>([^<]+)</td>\s*<td>([^<]+)</td>\s*<td>(\d{2}:\d{2})</td>\s*<td>(\d{2}:\d{2})</td>\s*<td>(\d{2}:\d{2})</td>\s*<td>(\d{2}:\d{2})</td>\s*<td>(\d{2}:\d{2})</td>\s*<td>(\d{2}:\d{2})</td>\s*</tr>`)

	matches := tableRowRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) >= 9 {
			tarih := strings.TrimSpace(match[1])
			hicriTarih := strings.TrimSpace(match[2])

			// ISO format tarih oluştur
			tarihISO := parseTarihToISO(tarih)

			gunlukVakitler := GunlukVakitler{
				Tarih:      tarih,
				TarihISO:   tarihISO,
				HicriTarih: hicriTarih,
				Vakitler: []NamazVakti{
					{VakitAdi: "İmsak", Vakit: match[3]},
					{VakitAdi: "Güneş", Vakit: match[4]},
					{VakitAdi: "Öğle", Vakit: match[5]},
					{VakitAdi: "İkindi", Vakit: match[6]},
					{VakitAdi: "Akşam", Vakit: match[7]},
					{VakitAdi: "Yatsı", Vakit: match[8]},
				},
			}

			vakitler = append(vakitler, gunlukVakitler)
		}
	}

	return vakitler
}

// Ana fonksiyon
func run(stateID, ilceID, vakitTipiStr, jsonFile string) error {

	// Türkiye verilerini yükle
	turkeyData, err := loadTurkeyData()
	if err != nil {
		return fmt.Errorf("Türkiye verileri yüklenemedi: %v", err)
	}

	// İl adını bul
	ilAdi, err := findIlNameByID(turkeyData, stateID)
	if err != nil {
		return fmt.Errorf("il bulunamadı: %v", err)
	}

	// İlçe listesini al
	ilceler, err := getIlceListesi(stateID)
	if err != nil {
		return fmt.Errorf("ilçe listesi alınamadı: %v", err)
	}

	if len(ilceler) == 0 {
		return fmt.Errorf("ilçe listesi boş")
	}

	// Vakit tipini belirle
	var vakitTipi VakitTipi
	switch strings.ToLower(vakitTipiStr) {
	case "gunluk":
		vakitTipi = Gunluk
	case "haftalik":
		vakitTipi = Haftalik
	case "yillik":
		vakitTipi = Yillik
	default:
		vakitTipi = Gunluk // varsayılan
	}

	// İlçe seçimi
	var secilenIlce Ilce
	if ilceID != "" {
		// Belirtilen ilçe ID'sini bul
		for _, ilce := range ilceler {
			if ilce.IlceID == ilceID {
				secilenIlce = ilce
				break
			}
		}
		if secilenIlce.IlceID == "" {
			return fmt.Errorf("belirtilen ilçe ID'si bulunamadı: %s", ilceID)
		}
	} else {
		// Varsayılan olarak merkez ilçeyi bul (il adıyla aynı olan)
		secilenIlce = ilceler[0] // fallback olarak ilk ilçe
		for _, ilce := range ilceler {
			if strings.EqualFold(ilce.IlceAdi, ilAdi) {
				secilenIlce = ilce
				break
			}
		}
	}

	html, err := getNamazVakitleriHTML(secilenIlce.IlceID)
	if err != nil {
		return fmt.Errorf("HTML alınamadı: %v", err)
	}

	// Sonuç yapısını oluştur
	sonuc := Sonuc{
		Il:        ilAdi,
		IlID:      stateID,
		Ilce:      secilenIlce.IlceAdi,
		IlceID:    secilenIlce.IlceID,
		VakitTipi: vakitTipi,
		Tarih:     time.Now().Format("2006-01-02"),
	}

	// Günlük vakitleri parse et
	gunlukVakitler := parseGunlukVakitler(html)

	if len(gunlukVakitler) == 0 {
		return fmt.Errorf("günlük vakitler parse edilemedi")
	}

	// Sonuçlara göre işlem yap
	switch vakitTipi {
	case Gunluk:
		sonuc.GunlukVakitler = gunlukVakitler

	case Haftalik:
		haftalikVakitler := parseNamazVakitleri(html)
		if len(haftalikVakitler) > 0 {
			// İlk 7 günü al
			buHafta := haftalikVakitler
			if len(buHafta) > 7 {
				buHafta = haftalikVakitler[:7]
			}
			sonuc.HaftalikVakitler = buHafta

			// İstatistikler
			sonuc.Istatistikler = map[string]interface{}{
				"toplam_gun": len(buHafta),
				"ilk_tarih":  buHafta[0].Tarih,
				"son_tarih":  buHafta[len(buHafta)-1].Tarih,
			}
		}

	case Yillik:
		yillikVakitler := parseNamazVakitleri(html)
		if len(yillikVakitler) > 0 {
			sonuc.YillikVakitler = yillikVakitler

			// İstatistikler
			sonuc.Istatistikler = map[string]interface{}{
				"toplam_gun": len(yillikVakitler),
				"ilk_tarih":  yillikVakitler[0].Tarih,
				"son_tarih":  yillikVakitler[len(yillikVakitler)-1].Tarih,
			}
		}
	}

	// JSON çıktısını stdout'a yazdır (minify)
	jsonData, err := json.Marshal(sonuc)
	if err != nil {
		return fmt.Errorf("JSON marshal hatası: %v", err)
	}
	fmt.Println(string(jsonData))

	// JSON dosyasına kaydet (eğer belirtilmişse)
	if jsonFile != "" {
		// Otomatik dosya yolu oluştur
		if jsonFile == "auto" {
			jsonFile = createFilePath(ilAdi, secilenIlce.IlceAdi, string(vakitTipi))
		}

		// Klasör oluştur
		if err := createDirectory(jsonFile); err != nil {
			return fmt.Errorf("klasör oluşturulamadı: %v", err)
		}

		if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
			return fmt.Errorf("JSON dosyası yazılamadı: %v", err)
		}
	}

	return nil
}

// İl listesini göster
func listIller() error {
	turkeyData, err := loadTurkeyData()
	if err != nil {
		return fmt.Errorf("Türkiye verileri yüklenemedi: %v", err)
	}

	// JSON çıktısını stdout'a yazdır (minify)
	jsonData, err := json.Marshal(turkeyData)
	if err != nil {
		return fmt.Errorf("JSON marshal hatası: %v", err)
	}
	fmt.Println(string(jsonData))

	return nil
}

func main() {
	var stateID string
	var ilceID string
	var vakitTipi string
	var jsonFile string

	rootCmd := &cobra.Command{
		Use:   "namaz-vakitleri",
		Short: "Diyanet Namaz Vakitleri CLI Uygulaması",
		Long: `Diyanet İşleri Başkanlığı'nın resmi web sitesinden namaz vakitlerini 
çeken ve görüntüleyen CLI uygulaması.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(stateID, ilceID, vakitTipi, jsonFile)
		},
	}

	// İl listesi komutu
	listCmd := &cobra.Command{
		Use:   "iller",
		Short: "Türkiye illerini listele",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listIller()
		},
	}

	rootCmd.AddCommand(listCmd)
	rootCmd.Flags().StringVarP(&stateID, "state", "s", "516", "İl ID'si (varsayılan: 516 - Bingöl)")
	rootCmd.Flags().StringVarP(&ilceID, "ilce", "i", "", "İlçe ID'si (belirtilmezse ilk ilçe kullanılır)")
	rootCmd.Flags().StringVarP(&vakitTipi, "vakit", "v", "gunluk", "Vakit tipi: gunluk, haftalik, yillik (varsayılan: gunluk)")
	rootCmd.Flags().StringVarP(&jsonFile, "json", "j", "", "JSON dosyasına kaydet (opsiyonel, 'auto' ile otomatik klasör yapısı)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(rootCmd.ErrOrStderr(), "Hata: %v\n", err)
	}
}

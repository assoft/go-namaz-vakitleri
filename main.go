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
func loadTurkeyData() (*TurkeyData, error) {
	data, err := os.ReadFile("data/turkey.json")
	if err != nil {
		return nil, fmt.Errorf("JSON dosyası okunamadı: %v", err)
	}

	var turkeyData TurkeyData
	if err := json.Unmarshal(data, &turkeyData); err != nil {
		return nil, fmt.Errorf("JSON parse hatası: %v", err)
	}

	return &turkeyData, nil
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
	fmt.Println("🕌 Diyanet Namaz Vakitleri Uygulaması")
	fmt.Println(strings.Repeat("=", 50))

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

	fmt.Printf("🏛️ Seçilen İl: %s (ID: %s)\n", ilAdi, stateID)

	// İlçe listesini al
	fmt.Println("📋 İlçeler alınıyor...")
	ilceler, err := getIlceListesi(stateID)
	if err != nil {
		return fmt.Errorf("ilçe listesi alınamadı: %v", err)
	}

	if len(ilceler) == 0 {
		return fmt.Errorf("ilçe listesi boş")
	}

	fmt.Printf("✅ %d ilçe bulundu:\n", len(ilceler))
	for i, ilce := range ilceler {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, ilce.IlceAdi, ilce.IlceID)
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
		// Varsayılan olarak ilk ilçe
		secilenIlce = ilceler[0]
	}

	fmt.Printf("\n🕐 %s için namaz vakitleri alınıyor...\n", secilenIlce.IlceAdi)

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

		// Konsol çıktısı
		fmt.Printf("\n📅 %s - Bugünün Namaz Vakitleri:\n", secilenIlce.IlceAdi)
		fmt.Println(strings.Repeat("=", 40))
		for _, vakit := range gunlukVakitler {
			fmt.Printf("   🕌 %s: %s\n", vakit.VakitAdi, vakit.Vakit)
		}

	case Haftalik:
		haftalikVakitler := parseNamazVakitleri(html)
		if len(haftalikVakitler) > 0 {
			// İlk 7 günü al
			buHafta := haftalikVakitler
			if len(buHafta) > 7 {
				buHafta = haftalikVakitler[:7]
			}
			sonuc.HaftalikVakitler = buHafta

			// Konsol çıktısı
			fmt.Printf("\n📆 %s - Bu Haftanın Namaz Vakitleri:\n", secilenIlce.IlceAdi)
			fmt.Println(strings.Repeat("=", 50))

			for _, gun := range buHafta {
				fmt.Printf("\n📅 %s (%s):\n", gun.Tarih, gun.HicriTarih)
				for _, vakit := range gun.Vakitler {
					fmt.Printf("   %s: %s\n", vakit.VakitAdi, vakit.Vakit)
				}
			}

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

			// Konsol çıktısı
			fmt.Printf("\n📅 %s - Yıllık Namaz Vakitleri:\n", secilenIlce.IlceAdi)
			fmt.Printf("✅ %d günlük vakit bulundu\n", len(yillikVakitler))

			// İstatistikler
			sonuc.Istatistikler = map[string]interface{}{
				"toplam_gun": len(yillikVakitler),
				"ilk_tarih":  yillikVakitler[0].Tarih,
				"son_tarih":  yillikVakitler[len(yillikVakitler)-1].Tarih,
			}

			fmt.Println("\n📊 İstatistikler:")
			fmt.Printf("   • Toplam gün sayısı: %d\n", len(yillikVakitler))
			fmt.Printf("   • İlk tarih: %s\n", yillikVakitler[0].Tarih)
			fmt.Printf("   • Son tarih: %s\n", yillikVakitler[len(yillikVakitler)-1].Tarih)
		}
	}

	// JSON dosyasına kaydet
	if jsonFile != "" {
		// Otomatik dosya yolu oluştur
		if jsonFile == "auto" {
			jsonFile = createFilePath(ilAdi, secilenIlce.IlceAdi, string(vakitTipi))
		}

		// Klasör oluştur
		if err := createDirectory(jsonFile); err != nil {
			return fmt.Errorf("klasör oluşturulamadı: %v", err)
		}

		jsonData, err := json.MarshalIndent(sonuc, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON marshal hatası: %v", err)
		}

		if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
			return fmt.Errorf("JSON dosyası yazılamadı: %v", err)
		}

		fmt.Printf("\n💾 Sonuçlar %s dosyasına kaydedildi\n", jsonFile)
	}

	return nil
}

// İl listesini göster
func listIller() error {
	turkeyData, err := loadTurkeyData()
	if err != nil {
		return fmt.Errorf("Türkiye verileri yüklenemedi: %v", err)
	}

	fmt.Println("🏛️ Türkiye İlleri:")
	fmt.Println(strings.Repeat("=", 50))

	for i, il := range turkeyData.Iller {
		fmt.Printf("%3d. %s (ID: %s)\n", i+1, il.Adi, il.ID)
	}

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

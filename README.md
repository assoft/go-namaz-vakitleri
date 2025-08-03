# Diyanet Namaz Vakitleri CLI

Diyanet İşleri Başkanlığı'nın resmi web sitesinden namaz vakitlerini çeken ve görüntüleyen Go CLI uygulaması.

## Özellikler

- 🕌 Günlük namaz vakitlerini görüntüleme
- 📅 Haftalık namaz vakitlerini görüntüleme
- 🏛️ İl ve ilçe seçimi
- 📊 İstatistikler
- 🎯 CLI parametreleri ile esnek kullanım

## Kurulum

### Gereksinimler

- Go 1.21 veya üzeri

### Kurulum Adımları

1. Projeyi klonlayın:
```bash
git clone <repo-url>
cd diyanet-namaz-vakitleri
```

2. Bağımlılıkları yükleyin:
```bash
go mod tidy
```

3. Uygulamayı derleyin:
```bash
go build -o namaz-vakitleri
```

## Kullanım

### Temel Kullanım

Varsayılan ayarlarla çalıştırma (Bingöl ili):
```bash
./namaz-vakitleri
```

### Parametreler

- `-s, --state`: İl ID'si (varsayılan: 516 - Bingöl)
- `-i, --ilce`: İlçe ID'si (belirtilmezse ilk ilçe kullanılır)
- `-v, --vakit`: Vakit tipi: gunluk, haftalik, yillik (varsayılan: gunluk)
- `-j, --json`: JSON dosyasına kaydet (opsiyonel, 'auto' ile otomatik klasör yapısı)

### Örnekler

Bingöl ili için (varsayılan):
```bash
./namaz-vakitleri
```

İl listesini görmek için:
```bash
./namaz-vakitleri iller
```

Farklı bir il için:
```bash
./namaz-vakitleri --state 539
```

Belirli bir ilçe için:
```bash
./namaz-vakitleri --state 539 --ilce 9541
```

Kısa parametrelerle:
```bash
./namaz-vakitleri -s 539 -i 9541
```

Farklı vakit tipleri:
```bash
# Günlük vakitler
./namaz-vakitleri -s 539 -v gunluk

# Haftalık vakitler
./namaz-vakitleri -s 539 -v haftalik

# Yıllık vakitler
./namaz-vakitleri -s 539 -v yillik
```

JSON çıktısı ile:
```bash
# Manuel dosya adı ile
./namaz-vakitleri -s 539 -v gunluk -j istanbul-gunluk.json

# Otomatik klasör yapısı ile
./namaz-vakitleri -s 539 -v gunluk -j auto
./namaz-vakitleri -s 539 -v haftalik -j auto
./namaz-vakitleri -s 539 -v yillik -j auto
```

## Çıktı Örneği

### Konsol Çıktısı

```
🕌 Diyanet Namaz Vakitleri Uygulaması
==================================================
🏛️ Seçilen İl: İSTANBUL (ID: 539)
📋 İlçeler alınıyor...
✅ 19 ilçe bulundu:
1. ARNAVUTKOY (ID: 9535)
2. AVCILAR (ID: 17865)
...

🕐 ARNAVUTKOY için namaz vakitleri alınıyor...

📅 ARNAVUTKOY - Bugünün Namaz Vakitleri:
========================================
   🕌 İmsak: 04:12
   🕌 Güneş: 05:55
   🕌 Öğle: 13:16
   🕌 İkindi: 17:10
   🕌 Akşam: 20:27
   🕌 Yatsı: 22:04

💾 Sonuçlar istanbul-gunluk.json dosyasına kaydedildi
```

### JSON Çıktısı

#### Günlük Vakitler
```json
{
  "il": "İSTANBUL",
  "il_id": "539",
  "ilce": "ARNAVUTKOY",
  "ilce_id": "9535",
  "vakit_tipi": "gunluk",
  "tarih": "2025-08-03",
  "gunluk_vakitler": [
    {
      "VakitAdi": "İmsak",
      "Vakit": "04:12"
    },
    {
      "VakitAdi": "Güneş",
      "Vakit": "05:55"
    },
    {
      "VakitAdi": "Öğle",
      "Vakit": "13:16"
    },
    {
      "VakitAdi": "İkindi",
      "Vakit": "17:10"
    },
    {
      "VakitAdi": "Akşam",
      "Vakit": "20:27"
    },
    {
      "VakitAdi": "Yatsı",
      "Vakit": "22:04"
    }
  ]
}
```

#### Haftalık/Yıllık Vakitler
```json
{
  "il": "İSTANBUL",
  "il_id": "539",
  "ilce": "ARNAVUTKOY",
  "ilce_id": "9535",
  "vakit_tipi": "haftalik",
  "tarih": "2025-08-03",
  "haftalik_vakitler": [
    {
      "tarih": "03 Ağustos 2025 Pazar",
      "tarih_iso": "2025-08-03",
      "hicriTarih": "9 Safer 1447",
      "vakitler": [
        {
          "VakitAdi": "İmsak",
          "Vakit": "04:12"
        },
        {
          "VakitAdi": "Güneş",
          "Vakit": "05:55"
        }
      ]
    }
  ],
  "istatistikler": {
    "toplam_gun": 7,
    "ilk_tarih": "03 Ağustos 2025 Pazar",
    "son_tarih": "09 Ağustos 2025 Cumartesi"
  }
}
```

## İl Kodları

Bazı popüler il kodları:
- 539: İstanbul
- 506: Ankara
- 540: İzmir
- 520: Bursa
- 507: Antalya
- 516: Bingöl (varsayılan)
- 500: Adana
- 532: Gaziantep
- 557: Mersin
- 566: Samsun
- 574: Trabzon

## Geliştirme

### Proje Yapısı

```
diyanet-namaz-vakitleri/
├── main.go          # Ana uygulama dosyası
├── go.mod           # Go modül dosyası
├── go.sum           # Bağımlılık hash'leri
├── data/
│   └── turkey.json  # Türkiye il verileri
├── vakitler/        # Otomatik oluşturulan JSON dosyaları
│   ├── istanbul/
│   │   ├── arnavutkoy/
│   │   │   └── gunluk.json
│   │   └── avcilar/
│   │       └── haftalik.json
│   └── ankara/
│       └── akyurt/
│           └── yillik.json
└── README.md        # Bu dosya
```

### Otomatik Klasör Yapısı

`-j auto` parametresi kullanıldığında dosyalar şu yapıda kaydedilir:

```
vakitler/
├── {il_adi}/
│   ├── {ilce_adi}/
│   │   ├── gunluk.json
│   │   ├── haftalik.json
│   │   └── yillik.json
│   └── {baska_ilce}/
│       └── ...
└── {baska_il}/
    └── ...
```

**Örnek:**
- `vakitler/istanbul/avcilar/haftalik.json`
- `vakitler/ankara/akyurt/yillik.json`
- `vakitler/bingol/merkez/gunluk.json`

**Not:** Türkçe karakterler otomatik olarak İngilizce karşılıklarına dönüştürülür:
- İ → I, ı → i
- Ğ → G, ğ → g
- Ü → U, ü → u
- Ş → S, ş → s
- Ö → O, ö → o
- Ç → C, ç → c
- Boşluklar ve özel karakterler → alt çizgi (_)

### Test

```bash
go test
```

### Lint

```bash
golangci-lint run
```

## Lisans

Bu proje MIT lisansı altında lisanslanmıştır.

## Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit yapın (`git commit -m 'Add some amazing feature'`)
4. Push yapın (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## API Desteği

Bu CLI uygulaması JSON çıktısı ürettiği için kolayca bir API'ye dönüştürülebilir:

### JSON Çıktı Formatı

```json
{
  "il": "İSTANBUL",
  "il_id": "539",
  "ilce": "ARNAVUTKOY",
  "ilce_id": "9535",
  "vakit_tipi": "gunluk|haftalik|yillik",
  "tarih": "2025-08-03",
  "gunluk_vakitler": [...],
  "haftalik_vakitler": [
    {
      "tarih": "03 Ağustos 2025 Pazar",
      "tarih_iso": "2025-08-03",
      "hicriTarih": "9 Safer 1447",
      "vakitler": [...]
    }
  ],
  "yillik_vakitler": [...],
  "istatistikler": {...}
}
```

### Tarih Formatları

- **tarih**: Türkçe format ("03 Ağustos 2025 Pazar")
- **tarih_iso**: ISO 8601 format ("2025-08-03") - API entegrasyonu için ideal
- **hicriTarih**: Hicri takvim format ("9 Safer 1447")

### API Entegrasyonu

JSON çıktısını kullanarak:
- Web API'leri oluşturabilirsiniz
- Mobil uygulamalar için veri sağlayabilirsiniz
- Veritabanına kaydedebilirsiniz
- Webhook'lar ile entegre edebilirsiniz

## Notlar

- Bu uygulama Diyanet İşleri Başkanlığı'nın resmi web sitesinden veri çeker
- API değişiklikleri uygulamanın çalışmasını etkileyebilir
- İnternet bağlantısı gereklidir
- JSON çıktısı UTF-8 encoding ile kaydedilir 
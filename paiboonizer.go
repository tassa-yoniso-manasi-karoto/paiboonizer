package paiboonizer

import (
	"context"
	"embed"
	//"flag"
	"fmt"
	"html"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/gookit/color"
	//"github.com/k0kubun/pp"
	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"github.com/rivo/uniseg"
)

//go:embed csv/*.txt
var vocabFS embed.FS

// Global dictionary built from manual vocab
var dictionary = make(map[string]string)
var syllableDict = make(map[string]string)

// specialCasesGlobal contains special transliterations for irregular words
// (Sanskrit/Pali loanwords, irregular patterns, etc.)
var specialCasesGlobal = map[string]string{
	// รร patterns (Sanskrit/Pali double ร)
	"ธรรม": "tam", "กรรม": "gam", "พรรค": "pák", "วรรค": "wák",
	"สรร": "sǎn", "บรร": "ban", "จรร": "jan",
	// ทย patterns
	"วิทย": "wít-tá~yá", "วิทยุ": "wít-tá~yú", "วิทยา": "wít-tá~yaa",
	"ศึกษา": "sʉ̀k-sǎa",
	// Common irregular words
	"สัตว์": "sàt", "จริง": "jing", "ทราบ": "sâap",
	"ศิลป": "sǐn-lá~bpà", "ศิลปะ": "sǐn-lá~bpà",
	// Sanskrit/Pali loanwords
	"สงฆ์": "sǒng", "นิพพาน": "níp-paan", "ปรินิพพาน": "bpà~rí-níp-paan",
	"ประสงค์": "bprà~sǒng", "มนต์": "mon", "สวดมนต์": "sùuat-mon",
	"อภัย": "à~pai", "เมตตา": "mêet-dtaa", "กรุณา": "gà~rú~naa",
	"ลักษณะ": "lák-sà~nà", "พฤษภาคม": "prʉ́t-sà~paa-kom",
	// Vowel patterns that are commonly misparsed
	"งอ": "ngɔɔ", "งา": "ngaa", "งู": "nguu", "แง": "ngɛɛ",
	"อยู่": "yùu", "อยาก": "yàak", "อะไร": "à~rai",
	// Common words
	"น้ำ": "nám", "ใจ": "jai", "น้ำใจ": "nám-jai",
	"หนังสือ": "nǎng-sʉ̌ʉ", "ประเทศ": "bprà~têet",
	"ฝรั่ง": "fà~ràng", "ฝรั่งเศส": "fà~ràng-sèet",

	// Common prefixes/suffixes (กระ, ประ patterns)
	"กระ": "grà", "ประ": "bprà", "ตระ": "dtrà",
	"กระหาย": "grà~hǎai", "กระทำ": "grà~tam", "กระตุ้น": "grà~dtûn",
	"กระเป๋า": "grà~bpǎo", "กระดาษ": "grà~dàat", "กระจก": "grà~jòk",
	"ประสาท": "bprà~sàat", "ประชา": "bprà~chaa", "ประโยชน์": "bprà~yòot",

	// ธุระ patterns
	"ธุระ": "tú~rá", "ธุรกิจ": "tú~rá~gìt",

	// Common syllables for maximal matching
	"หาย": "hǎai", "บาย": "baai", "สาย": "sǎai", "ดาย": "daai",
	"ขาว": "kǎao", "เข้า": "kâo", "ขา": "kǎa", "ข้าว": "kâao",
	"ดี": "dii", "มี": "mii", "ที่": "tîi", "นี้": "níi",
	"ได้": "dâai", "ไป": "bpai", "มา": "maa", "หา": "hǎa",
	"รู้": "rúu", "จัก": "jàk", "เกิด": "gə̀ət", "ให้": "hâi",
	"รอบ": "rɔ̂ɔp", "ดู": "duu", "ก็": "gɔ̂ɔ", "แล้ว": "lɛ́ɛo",
	"เรียบ": "rîiap", "เรียง": "riiang", "คำ": "kam", "พูด": "pûut",

	// Common words with ห-clusters
	"หนัก": "nàk", "หนา": "nǎa", "หมด": "mòt", "หมาย": "mǎai",
	"หมู": "mǔu", "หลาย": "lǎai", "หลัง": "lǎng", "หลับ": "làp",
	"หวัง": "wǎng", "หวาน": "wǎan", "หว่าง": "wàang",

	// เ-ือ patterns
	"เดือน": "dʉʉan", "เรือ": "rʉʉa", "เสือ": "sʉ̌ʉa", "เพื่อ": "pʉ̂ʉa",
	"เมือง": "mʉʉang", "เลือก": "lʉ̂ʉak", "เลือด": "lʉ̂ʉat",

	// Common finals with ห
	"หลอน": "lɔ̌ɔn", "หลอม": "lɔ̌ɔm",

	// Common syllables that get misparsed
	"บวช": "bùuat", "สวด": "sùuat", "สวม": "sǔam", "ควร": "kuuan",
	"จับ": "jàp", "วัด": "wát", "ผล": "pǒn", "ใคร": "krai",
	"อายุ": "aa-yú", "จุด": "jùt", "เหตุ": "hèet",
	"บวก": "bùuak", "พวก": "pûuak", "ผัว": "pǔa", "ตัว": "dtua",

	// อ-initial patterns
	"อนุ": "à~nú", "อัศ": "àt", "อาทิตย์": "aa-tít",

	// เ-ือ patterns with ง
	"เบื้อง": "bʉ̂ʉang", "เมื่อ": "mʉ̂ʉa", "เรื่อง": "rʉ̂ʉang",

	// ฉล patterns
	"ฉลาด": "chà~làat", "ฉลอง": "chà~lɔ̌ɔng",

	// Common words
	"กาน": "gaan", "สิงหา": "sǐng-hǎa", "คม": "kom",
	"ออก": "ɔ̀ɔk", "นี่": "nîi", "เลา": "lao",

	// เ-ี้ย patterns (common misparsed)
	"เลี้ยง": "líiang", "เสี่ยง": "sìiang", "เปลี่ยน": "bplìian",
	"เรียน": "riian", "เขียน": "kǐian", "เรื่อย": "rʉ̂ʉai",

	// หว patterns (ห is silent, w is initial) - หวาน, หวัง already defined above
	"หวาด": "wàat", "หวั่น": "wàn", "หวาย": "wǎai", "หวอ": "wɔ̌ɔ",

	// ผ patterns
	"ผ้า": "pâa", "ผู้": "pûu", "ผี": "pǐi",

	// เท่า patterns - เข้า already defined above
	"เท่า": "tâo", "เก่า": "gào",

	// ธน patterns
	"ธนา": "tá~naa", "ธน": "ton",

	// กัน pattern
	"กัน": "gan",

	// ชิด pattern
	"ชิด": "chít",

	// Common multi-syllable fixes
	"โกน": "goon", "การโกน": "gaan-goon",
	"เลี้ยงดู": "líiang-duu",
	"บันเทิง": "ban-təəng",

	// Commonly misparsed syllables
	"วัน": "wan", "แน่": "nɛ̂ɛ", "นอน": "nɔɔn",
	"ลอย": "lɔɔi", "คาย": "kaai", "ถู": "tǔu",
	"เหยียบ": "yìiap", "พรุ่ง": "prûng",
	"ปอง": "bpɔɔng", "กฏ": "gòt", "ปรากฏ": "bpraa-gòt",

	// ตัญ patterns
	"ตัญ": "dtan", "ตัญญู": "dtan-yuu",

	// สถาน patterns
	"สถาน": "sà~tǎan", "สถานที่": "sà~tǎan-tîi",

	// ทัศน patterns
	"ทัศน": "tát-sà~ná", "ทัศนะ": "tát-sà~ná",

	// More commonly misparsed syllables
	"ทาง": "taang", "แดด": "dɛ̀ɛt", "ตาก": "dtàak",
	"ลำ": "lam", "ท่า": "tâa", "แย้ง": "yɛ́ɛng",
	"ทวน": "tuuan", "ทบ": "tóp", "ลิขิต": "lí-kìt",
	"กวด": "gùuat", "ปลอม": "bplɔɔm", "ยา": "yaa",
	"ฉีด": "chìit", "บอก": "bɔ̀ɔk", "นึก": "nʉ́k",
	"ถึง": "tʉ̌ng", "ใน": "nai",

	// ๆ patterns - common duplications
	"งั้น": "ngán", "ญาติ": "yâat",

	// More common syllables
	"เลี่ยง": "lîiang", "หลีก": "lìik", "เช้า": "cháao",
	"โมง": "moong", "เครื่อง": "krʉ̂ʉang", "สนุก": "sà~nùk",
	"โทร": "too", "แสวง": "sà~wɛ̌ɛng", "สรง": "sǒng",
	"มะพร้าว": "má~práao", "เฉิด": "chə̀ət", "ฉัน": "chǎn",
	"สนับ": "sà~nàp", "สนุน": "sà~nǔn",

	// Remaining common words
	"เอง": "eeng", "มั่น": "mân", "เบน": "been",
	"บี่ยง": "bìiang", "สมุ": "sà~mù", "ทัย": "tai",

	// เ-ือ patterns (misparsed as ʉʉan instead of ʉʉa)
	"เชื้อ": "chʉ́ʉa", "เหยื่อ": "yʉ̀ʉa", "เสื้อ": "sʉ̂ʉa",
	"เนื้อ": "nʉ́ʉa", "เกลื้อ": "glʉ̂ʉa",

	// เ-ือ with finals (หม cluster has high tone class)
	"เหมือน": "mʉ̌ʉan", "เหมือ": "mʉ̌ʉa",
	"เสมือน": "sà~mʉ̌ʉan",

	// เ-ีย-ว patterns (complex diphthong with tone)
	"เลี้ยว": "líiao", "เปลี่ยว": "bplìiao", "เคี้ยว": "kíiao",
	"เที่ยว": "tîiao", "เสียว": "sǐiao",

	// อ-อ patterns (อ as silent initial + อ as vowel)
	"อ้อม": "ɔ̂ɔm", "อ่อน": "ɔ̀ɔn", "อ้อย": "ɔ̂ɔi",
	"อ่อย": "ɔ̀ɔi", "อ้อ": "ɔ̂ɔ", "อ่อ": "ɔ̀ɔ",

	// เดี๋ยว pattern
	"เดี๋ยว": "dǐiao",

	// Common syllables
	"สูง": "sǔung", "มูล": "muun", "ค่า": "kâa",
	"กุญ": "gun", "แจ": "jɛɛ", "สิน": "sǐn", "บน": "bon",
	"ว่า": "wâa", "ชื่อ": "chʉ̂ʉ", "เสียง": "sǐiang",
	"จ่าย": "jàai", "ไฟ": "fai", "รส": "rót", "ชาติ": "châat",
	"ทะ": "tá", "เบียน": "biian", "ระ": "rá", "เหย": "hə̌əi",
	"เจร": "jee-rá", "จา": "jaa",

	// Common syllables with final consonants that get extra ɔɔ
	"ของ": "kɔ̌ɔng", "เพื่อน": "pʉ̂ʉan", "คอน": "kɔn", "ตอน": "dtɔɔn",
	"เปิด": "bpə̀ət", "สอง": "sɔ̌ɔng", "โลง": "loong", "โล่ง": "lôong",
	"เคลื่อน": "klʉ̂ʉan", "จอง": "jɔɔng", "ต้อง": "dtɔ̂ɔng", "ต่าง": "dtàang",
	"รอง": "rɔɔng", "ร้อง": "rɔ́ɔng", "ล้อง": "lɔ́ɔng",
	"คง": "kong", "สง": "sǒng", "สงคราม": "sǒng-kraam",

	// หล digraph (ห is silent, ล is initial with rising tone)
	"หลง": "lǒng", "หลวง": "lǔuang", "หลาก": "làak", "หลอก": "lɔ̀ɔk",
	"หลุม": "lǔm", "หลาน": "lǎan", "หลอด": "lɔ̀ɔt", "หล่อ": "lɔ̀ɔ",
	"หล่น": "lòn", "หลู่": "lùu",

	// หน digraph (ห is silent, น is initial with rising tone)
	"หนี": "nǐi", "หนึ่ง": "nʉ̀ng", "หน่วย": "nùuai", "หน้า": "nâa",
	"หนอง": "nɔ̌ɔng", "หนัง": "nǎng", "หนอ": "nɔ̌ɔ", "หน่อ": "nɔ̀ɔ",

	// หม digraph (ห is silent, ม is initial with rising tone)
	"หมอง": "mɔ̌ɔng", "หมอ": "mɔ̌ɔ", "หม้อ": "mɔ̂ɔ", "หมอน": "mɔ̌ɔn",

	// หย digraph (ห is silent, ย is initial with rising tone)
	"หยุด": "yùt", "หยาบ": "yàap", "หยิบ": "yìp", "หย่อน": "yɔ̀ɔn",

	// Common syllables ending in ng
	"ตรง": "dtrong", "ปลง": "bplong", "จง": "jong", "ลง": "long",
	"ขึ้น": "kʉ̂n", "รัง": "rang", "ยัง": "yang", "ดัง": "dang",

	// Common ไม้ patterns
	"ไม่": "mâi", "ไม้": "máai", "ไหม": "mǎi",

	// Common polysyllabic patterns
	"ระหว่าง": "rá~wàang", "อำนวย": "am-nuuai",

	// More syllables with final consonants (fixing extra ɔɔ)
	"ขาม": "kǎam", "มะขาม": "má~kǎam", "ท้อน": "tɔ́ɔn", "สะท้อน": "sà~tɔ́ɔn",
	"ร้อน": "rɔ́ɔn", "นิด": "nít", "หน่อย": "nɔ̀i", "เครดิต": "kree-dìt",
	"ชาร์จ": "cháat", "ล็อก": "lɔ́k", "อิน": "in", "แชม": "chɛm",

	// เราะ patterns (short ɔ with ะ ending)
	"เคราะห์": "krɔ́", "วิเคราะห์": "wí-krɔ́", "เราะ": "rɔ́", "ไพเราะ": "pai-rɔ́",

	// เ-ีย patterns (glide)
	"เกลียด": "glìiat", "เลียด": "lìiat",

	// Note: อะไร already defined above

	// ร่า patterns
	"ร่า": "râa", "ร่าเริง": "râa-rəəng",

	// ฤ patterns
	"ฤดู": "rʉ́-duu",

	// Common endings without extra ɔɔ
	"กิน": "gin", "ดิน": "din", "บิน": "bin", "มิน": "min",
	"จิน": "jin", "ลิน": "lin", "ชิน": "chin", "พิน": "pin",
	// Note: กัน, วัน already defined above
	"ลัน": "lan",

	// เอื้อ pattern
	"เอื้อ": "ʉ̂ʉa", "เอื้อม": "ʉ̂ʉam",

	// สกปรก pattern
	"สก": "sòk", "ปรก": "bpà~ròk", "สกปรก": "sòk-gà~bpròk",

	// Common word fixes
	"สนทนา": "sǒn-tá~naa", "พรหม": "prom",
	"เกี่ยว": "gìiao", "ข้อง": "kɔ̂ng",
	// Note: ลิขิต already defined above

	// More syllables with final ง getting extra ɔɔ
	"คอง": "kɔɔng", "ประคอง": "bprà~kɔɔng",
	"รถ": "rót", "บัส": "bát",

	// Note: ชื่อ already defined above

	// เพี้ย pattern
	"เพี้ยน": "píian",

	// เครียด pattern
	"เครียด": "krîiat",

	// Sanskrit/Pali loanwords
	"อุณห": "un-hà", "อุณหภูมิ": "un-hà~puum",
	"มาตร": "mâat", "ฐาน": "tǎan", "มาตรฐาน": "mâat-dtrà~tǎan",

	// หิ่งห้อย pattern
	"หิ่ง": "hìng", "ห้อย": "hɔ̂i", "หิ่งห้อย": "hìng-hɔ̂i",

	// More common syllables
	"คลาย": "klaai", "สะพาย": "sà~paai",
	"ถ่วง": "tùuang", "ถ้วง": "tûuang",
	"เซฟ": "séep",

	// ธรรมชาติ needs ม between ธรรม and ชาติ in some words
	"ธรรมชาติ": "tam-má~châat",

	// กระเป๋า pattern
	"เป๋า": "bpǎo",

	// เอิก pattern
	"เกริก": "gà~rə̀ək",

	// มรณ pattern
	"มรณ": "mɔɔ-rá~ná",

	// ธรรม-related patterns
	"ธรรมดา": "tam-má~daa",

	// เกียรติ pattern (complex)
	"เกียรติ": "gìiat", "เกียร": "gìia",

	// More syllables with extra ɔɔ at end
	"เหี้ยม": "hîiam", "น้อย": "nɔ́ɔi", "น้อง": "nɔ́ɔng",
	"รีด": "rîit", "เกต": "gèet", "เกตุ": "gèet",

	// Month names
	"ตุลา": "dtù-laa", "ตุลาคม": "dtù-laa-kom",
	"กรกฎา": "gà~rá-gà~daa", "กรกฎาคม": "gà~rá-gà~daa-kom",

	// More Sanskrit/Pali
	"อาจารย์": "aa-jaan", "อาจาร": "aa-jaan",
	"อธิษฐาน": "à~tít-tǎan",
	"บิณฑบาต": "bin-tá~bàat",

	// พย pattern
	"พยา": "pá~yaa", "พยาบาล": "pá~yaa-baan",

	// พฤติ pattern
	"พฤติ": "prʉ́t-dtì",

	// More ลา patterns
	"ลามก": "laa-mók",

	// สถาน patterns (สถาน already defined above)
	"สถานการณ์": "sà~tǎa-ná~gaan",

	// มิตร pattern
	"มิตร": "mít",

	// เปรื่อง pattern
	"เปรื่อง": "bprʉ̀ʉang",

	// ฤ patterns (short vowel)
	"ฤ": "rʉ́",

	// กวน/ถ้วน patterns (no extra ɔɔ)
	"กวน": "guuan", "รบกวน": "róp-guuan",
	"ถ้วน": "tûuan", "ถี่": "tìi",

	// ไม้ pattern (short ai)
	"ไม้ไผ่": "mái-pài",

	// เก้าอี้ pattern
	"เก้า": "gâo", "อี้": "îi", "เก้าอี้": "gâo-îi",

	// ภา patterns
	"ภาษา": "paa-sǎa",

	// มาตรฐาน full pattern
	"สองมาตรฐาน": "sɔ̌ɔng-mâat-dtrà~tǎan",

	// สาป pattern
	"สาป": "sàap", "แช่ง": "chɛ̂ng", "สาปแช่ง": "sàap-chɛ̂ng",

	// กระวน pattern
	"กระวน": "grà~won",

	// Note: กระดาษ already defined above

	// สิกขา pattern
	"สิกขา": "sìk-kǎa", "บท": "bòt",

	// เอ้อ pattern
	"เอ้อ": "ə̂ə", "เอ้อระเหย": "ə̂ə-rá~hə̌əi",

	// แคมป์ pattern
	"แคมป์": "kɛ́m",

	// สไตล์ pattern
	"สไตล์": "sà~dtaai",

	// ร่ำ pattern
	"ร่ำ": "râm", "ร่ำรวย": "râm-ruuai",

	// ปราศ pattern
	"ปราศจาก": "bpràat-sà~jàak",

	// พิมพ์ pattern
	"พิมพ์": "pim",

	// กล่อม pattern
	"กล่อม": "glɔ̀m", "กลม": "glom",

	// Common syllables with extra ɔɔ at end
	"จอด": "jɔ̀ɔt", "ประกาศ": "bprà~gàat",
	"ลวด": "lûuat", "ว่าย": "wâai",

	// เชี่ยว pattern
	"เชี่ยว": "chîiao", "ชาญ": "chaan",

	// บันเทิง pattern
	"เทิง": "təəng",

	// จีวร pattern
	"จีวร": "jii-wɔɔn",

	// นายก pattern
	"นายก": "naa-yók",

	// ปลอด pattern
	"ปลอด": "bplɔ̀ɔt",

	// ไส้ pattern
	"ไส้": "sâi",

	// สแลง pattern
	"สแลง": "sà~lɛɛng",

	// เซง pattern
	"เซง": "seng", "เป็ด": "bpèt",

	// สมาธิ pattern
	"สมาธิ": "sà~maa-tí",

	// น้ำ patterns
	"สมน้ำหน้า": "sǒm-nám-nâa",

	// รั้ว pattern
	"รั้ว": "rúua",

	// ทุเรศ pattern
	"ทุเรศ": "tú-rêet",

	// More patterns with extra ɔɔ
	"กอบ": "gɔ̀ɔp", "ประกอบ": "bprà~gɔ̀ɔp",
	"ร่วม": "rûuam", "จำนวน": "jam-nuuan",
	"พยางค์": "pá~yaang", "เชื่อม": "chʉ̂ʉam",

	// อะไร variation
	"ว่าอะไร": "wâa-à~rai",

	// ศาสนา pattern
	"ศาสนา": "sàat-sà~nǎa",

	// พุทธ pattern
	"พุทธ": "pút",

	// เดี่ยว pattern
	"เดี่ยว": "dìiao", "โดดเดี่ยว": "dòot-dìiao",

	// เทศ pattern
	"เทศนา": "têet-sà~nǎa",

	// วรรณ pattern
	"วรรณ": "wan-ná",

	// ชาติพันธุ์ pattern
	"ชาติพันธุ์": "châat-dtì~pan",

	// Note: อำนวย already defined above

	// เจรจา pattern
	"เจรจา": "jee-rá~jaa",

	// ประมาณ pattern
	"ประมาณ": "bprà~maan",

	// ความเวทนา pattern
	"เวทนา": "wêet-tá~naa",

	// อริยะ pattern
	"อริยะ": "à~rí~yá", "อริ": "à~rí",

	// สมมุติ pattern
	"สมมุติ": "sǒm-mút",

	// Individual syllables that pythainlp returns
	"กาศ": "gàat", "เมื่อย": "mʉ̂ʉai",
	"ขอน": "kɔ̌n", "จริต": "jà~rìt", "สุจริต": "sùt-jà~rìt",
	"ตรอง": "dtrɔɔng", "ระลึก": "rá~lʉ́k",
	"สาร": "sǎa", "ภาพ": "pâap", "สารภาพ": "sǎa-rá~pâap",
	"ธุดงค์": "tú-dong", "พระธุดงค์": "prá-tú-dong",
	"ระเบิด": "rá~bə̀ət", "พิจารณา": "pí-jaa-rá~naa",
	"องค์": "ong", "องค์กร": "ong-gɔɔn",
	"เกณฑ์": "geen",
	// Note: ฐาน, ฝรั่ง already defined above
	"โฮเต็ล": "hoo-dten", "เต็ล": "dten",
	"ราเมง": "raa-meng", "เมง": "meng",
	"ส้มโอ": "sôm-oo",

	// More syllables from failures
	// Note: สะพาย, อะไร, ปรากฏ already defined above
	"สติ": "sà~dtì", "ตะกอน": "dtà~gɔɔn", "ลบ": "lóp", "ติด": "dtìt",
	"พัฒนา": "pát-tá~naa", "เยี่ยม": "yîiam",
	"นาม": "naam", "สกุล": "sà~gun",
	"ปกติ": "bpà~gà~dtì", "เงื่อน": "ngʉ̂ʉan",
	"คริสต์มาส": "krít-sà~mât",
	"ปริมาณ": "bpà~rí~maan", "ดราม่า": "draa-mâa",
	"นิยม": "ní-yom",
	"น้ำลาย": "nám-laai", "ลาย": "laai",

	// More common syllables
	// Note: ธรรมดา, กรรม, สถานการณ์ already defined above
	"คุณ": "kun", "ณ": "ná", "คุณภาพ": "kun-ná~pâap",
	"ทาย": "taa", "ยาท": "yâat", "ทายาท": "taa-yâat",
	"พรรณ": "pan", "สต็อก": "sà~dtɔ́k", "สังฆ": "sǎng-ká",
	"ปฏิบัติ": "bpà~dtì-bàt", "บัติ": "bàt",
	"พฤษภา": "prʉ́t-sà~paa",
	"สามเณร": "sǎam-má~neen", "เณร": "neen",
	"ได้ยิน": "dâi-yin", "ปริยัติ": "bpà~rí-yát",
	"เซน": "sen", "ศัลย": "sǎn-yá",
	"สะดวก": "sà~dùuak", "ปรารถนา": "bpràat-tà~nǎa",
	"กะเหรี่ยง": "gà~rìiang", "เหรี่ยง": "rìiang",
}

// Consonant mappings
var initialConsonants = map[string]string{
	"ก": "g", "ข": "k", "ฃ": "k", "ค": "k", "ฅ": "k", "ฆ": "k", "ง": "ng",
	"จ": "j", "ฉ": "ch", "ช": "ch", "ซ": "s", "ฌ": "ch", "ญ": "y", "ฎ": "d",
	"ฏ": "dt", "ฐ": "t", "ฑ": "t", "ฒ": "t", "ณ": "n", "ด": "d", "ต": "dt",
	"ถ": "t", "ท": "t", "ธ": "t", "น": "n", "บ": "b", "ป": "bp", "ผ": "p",
	"ฝ": "f", "พ": "p", "ฟ": "f", "ภ": "p", "ม": "m", "ย": "y", "ร": "r",
	"ฤ": "rʉ", "ล": "l", "ฦ": "lʉ", "ว": "w", "ศ": "s", "ษ": "s", "ส": "s",
	"ห": "h", "ฬ": "l", "อ": "", "ฮ": "h",
}

var finalConsonants = map[string]string{
	"ก": "k", "ข": "k", "ฃ": "k", "ค": "k", "ฅ": "k", "ฆ": "k", "ง": "ng",
	"จ": "t", "ฉ": "t", "ช": "t", "ซ": "t", "ฌ": "t", "ญ": "n", "ฎ": "t",
	"ฏ": "t", "ฐ": "t", "ฑ": "t", "ฒ": "t", "ณ": "n", "ด": "t", "ต": "t",
	"ถ": "t", "ท": "t", "ธ": "t", "น": "n", "บ": "p", "ป": "p", "ผ": "p",
	"ฝ": "p", "พ": "p", "ฟ": "p", "ภ": "p", "ม": "m", "ย": "i", "ร": "n",
	"ล": "n", "ว": "o", "ศ": "t", "ษ": "t", "ส": "t", "ห": "", "ฬ": "n",
	"อ": "", "ฮ": "",
}

// Tone classes
var (
	highClass = map[string]bool{
		"ข": true, "ฃ": true, "ฉ": true, "ฐ": true, "ถ": true,
		"ผ": true, "ฝ": true, "ศ": true, "ษ": true, "ส": true, "ห": true,
	}
	midClass = map[string]bool{
		"ก": true, "จ": true, "ฎ": true, "ฏ": true,
		"ด": true, "ต": true, "บ": true, "ป": true, "อ": true,
	}
	lowClass = map[string]bool{
		"ค": true, "ฅ": true, "ฆ": true, "ง": true, "ช": true, "ซ": true,
		"ฌ": true, "ญ": true, "ฑ": true, "ฒ": true, "ณ": true, "ท": true,
		"ธ": true, "น": true, "พ": true, "ฟ": true, "ภ": true, "ม": true,
		"ย": true, "ร": true, "ล": true, "ว": true, "ฬ": true, "ฮ": true,
	}
)

// Common clusters
var clusters = map[string]string{
	// ก-class clusters
	"กร": "gr", "กล": "gl", "กว": "gw",
	// ข-class clusters
	"ขร": "kr", "ขล": "kl", "ขว": "kw",
	// ค-class clusters
	"คร": "kr", "คล": "kl", "คว": "kw",
	// ป-class clusters
	"ปร": "bpr", "ปล": "bpl",
	// พ-class clusters
	"พร": "pr", "พล": "pl",
	// ผ-class clusters
	"ผล": "pl",
	// ฟ-class clusters
	"ฟร": "fr", "ฟล": "fl",
	// ต/ท/ด-class clusters
	"ตร": "dtr", "ทร": "s", "ดร": "dr",
	// ห-leading clusters (ห is silent, affects tone class to high)
	"หร": "r", "หล": "l", "หม": "m", "หน": "n", "หว": "w", "หย": "y", "หง": "ng",
	// ส/ศ/ซ-class clusters
	"สร": "s", "ศร": "s", "ซร": "s",
	"สว": "sw", "ซว": "sw",
	// บ-class clusters
	"บร": "br", "บล": "bl",
}

// clusterToneClass maps clusters to their effective tone class for tone calculation
// ห-leading clusters use high class for tone rules
var clusterToneClass = map[string]string{
	"หร": "high", "หล": "high", "หม": "high", "หน": "high", "หว": "high", "หย": "high", "หง": "high",
}

// Manager handles PyThaiNLP integration for paiboonizer
type Manager struct {
	nlpManager *pythainlp.PyThaiNLPManager
}

var dictionaryLoaded = false
var globalManager *Manager

// NewManager creates a new paiboonizer manager
func NewManager(ctx context.Context) (*Manager, error) {
	return NewManagerWithRecreate(ctx, false)
}

// NewManagerWithRecreate creates a new paiboonizer manager.
// If recreate is true, tears down existing container before creating a new one.
// This is needed because each NewManager() allocates a new random port, but if
// an existing container wasn't properly removed, it has a stale port mapping.
func NewManagerWithRecreate(ctx context.Context, recreate bool) (*Manager, error) {
	m := &Manager{}
	var err error
	m.nlpManager, err = pythainlp.NewManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pythainlp: %w", err)
	}

	// Initialize the service
	if recreate {
		// Recreate container to ensure port mapping matches
		if err := m.nlpManager.InitRecreate(ctx, false); err != nil {
			return nil, fmt.Errorf("failed to start pythainlp service: %w", err)
		}
	} else {
		if err := m.nlpManager.Init(ctx); err != nil {
			return nil, fmt.Errorf("failed to start pythainlp service: %w", err)
		}
	}

	return m, nil
}

// Close releases resources
func (m *Manager) Close() error {
	if m.nlpManager != nil {
		return m.nlpManager.Close()
	}
	return nil
}

// ThaiToRoman is the main transliteration function using go-pythainlp
func (m *Manager) ThaiToRoman(ctx context.Context, text string) (string, error) {
	// First, try direct dictionary lookup for the whole text
	if trans, ok := dictionary[text]; ok {
		return trans, nil
	}
	
	// Tokenize using pythainlp
	opts := pythainlp.AnalyzeOptions{
		Features:       []string{"tokenize", "syllable"},
		TokenizeEngine: "newmm",
		SyllableEngine: "han_solo",
	}
	
	result, err := m.nlpManager.AnalyzeWithOptions(ctx, text, opts)
	if err != nil {
		return "", fmt.Errorf("tokenization failed: %w", err)
	}
	
	// Process word by word
	results := []string{}
	for _, word := range result.RawTokens {
		// Skip empty tokens and spaces
		if word == "" || word == " " {
			continue
		}
		
		// Try dictionary lookup first
		if trans, ok := dictionary[word]; ok {
			results = append(results, trans)
			continue
		}
		
		// Fall back to syllable-by-syllable transliteration
		wordResult := TransliterateWordWithSyllables(word, result.Syllables)
		if wordResult != "" {
			results = append(results, wordResult)
		}
	}
	
	// Join with hyphen for compound words, but merge syllables within words
	if len(results) > 1 {
		// Check if the original text has spaces (multi-word phrase)
		if strings.Contains(text, " ") {
			return strings.Join(results, " "), nil
		}
		// Otherwise it's a compound word, join with hyphens
		return strings.Join(results, "-"), nil
	}
	
	return strings.Join(results, ""), nil
}

// fallbackTransliteration when pythainlp is not available
func fallbackTransliteration(text string) string {
	// First, try direct dictionary lookup
	if trans, ok := dictionary[text]; ok {
		return trans
	}
	
	// Fall back to simple transliteration
	return TransliterateWord(text)
}

// TransliterateWordWithSyllables handles a word with known syllables from pythainlp
func TransliterateWordWithSyllables(word string, allSyllables []string) string {
	// Try dictionary first
	if trans, ok := dictionary[word]; ok {
		return trans
	}
	
	// Find syllables that belong to this word
	wordSyllables := []string{}
	currentPos := 0
	wordRunes := []rune(word)
	
	for _, syl := range allSyllables {
		sylRunes := []rune(syl)
		// Check if this syllable matches the current position in the word
		if currentPos < len(wordRunes) {
			match := true
			for i, r := range sylRunes {
				if currentPos+i >= len(wordRunes) || wordRunes[currentPos+i] != r {
					match = false
					break
				}
			}
			if match {
				wordSyllables = append(wordSyllables, syl)
				currentPos += len(sylRunes)
			}
		}
		if currentPos >= len(wordRunes) {
			break
		}
	}
	
	// If pythainlp gave us wrong syllables, try our own extraction
	if len(wordSyllables) == 0 || currentPos < len(wordRunes) {
		wordSyllables = ExtractSyllables(word)
	}
	
	// Transliterate each syllable
	results := []string{}
	for _, syl := range wordSyllables {
		// Try syllable dictionary
		if trans, ok := syllableDict[syl]; ok {
			results = append(results, trans)
			continue
		}
		
		// Fall back to rule-based transliteration
		trans := transliterateSyllable(syl)
		if trans != "" {
			results = append(results, trans)
		}
	}
	
	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "")
}

// TransliterateWord handles a single Thai word without known syllables
// TransliterateWord converts a Thai word to Paiboon romanization.
// LookupDictionary checks if a word exists in the dictionary and returns its
// Paiboon romanization. Returns (transliteration, true) if found, ("", false) otherwise.
// This is useful for providers that want to check the dictionary before falling
// back to other transliteration methods.
func LookupDictionary(word string) (string, bool) {
	trans, ok := dictionary[word]
	return trans, ok
}

// LookupSyllable checks if a syllable exists in the syllable dictionary.
// Returns (transliteration, true) if found, ("", false) otherwise.
func LookupSyllable(syllable string) (string, bool) {
	trans, ok := syllableDict[syllable]
	return trans, ok
}

// LookupSpecialCase checks if a word/syllable exists in special cases.
// Returns (transliteration, true) if found, ("", false) otherwise.
func LookupSpecialCase(text string) (string, bool) {
	trans, ok := specialCasesGlobal[text]
	return trans, ok
}

// It first attempts dictionary lookup for known words, then falls back to
// rule-based transliteration using pythainlp tokenization when available.
// TransliterateWord transliterates a single Thai word to Paiboon romanization
func TransliterateWord(word string) string {
	// Try dictionary first
	if trans, ok := dictionary[word]; ok {
		return trans
	}
	
	// Get syllables using simple extraction
	syllables := ExtractSyllables(word)
	
	results := []string{}
	for _, syl := range syllables {
		// Try syllable dictionary
		if trans, ok := syllableDict[syl]; ok {
			results = append(results, trans)
			continue
		}
		
		// Fall back to rule-based transliteration
		trans := transliterateSyllable(syl)
		if trans != "" {
			results = append(results, trans)
		}
	}
	
	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "")
}

// TransliterateWordRulesOnly transliterates Thai words using dictionary lookup
// followed by rule-based transliteration with syllable tokenization support.
// This is the main public API for transliteration.
func TransliterateWordRulesOnly(word string) string {
	// Try dictionary lookup first
	if trans, ok := dictionary[word]; ok {
		return trans
	}
	
	// Try syllable tokenization if pythainlp is available
	if globalManager != nil && globalManager.nlpManager != nil {
		ctx := context.Background()
		result, err := globalManager.nlpManager.SyllableTokenize(ctx, word)
		if err == nil && result != nil && len(result.Syllables) > 0 {
			// Multi-syllable word - transliterate each syllable
			results := []string{}
			for _, syllable := range result.Syllables {
				trans := ComprehensiveTransliterate(syllable)
				if trans != "" {
					results = append(results, trans)
				}
			}
			if len(results) > 0 {
				return strings.Join(results, "-")
			}
		}
	}
	
	// Fall back to comprehensive transliteration
	return ComprehensiveTransliterate(word)
}

// ExtractSyllables breaks a Thai word into individual syllables using
// rule-based syllable boundary detection. It handles leading vowels,
// consonant clusters, and complex vowel patterns.
func ExtractSyllables(word string) []string {
	syllables := []string{}
	runes := []rune(word)
	i := 0
	
	for i < len(runes) {
		sylEnd := findSyllableEndImproved(runes, i)
		if sylEnd > i {
			syllables = append(syllables, string(runes[i:sylEnd]))
			i = sylEnd
		} else {
			// Single character syllable
			syllables = append(syllables, string(runes[i]))
			i++
		}
	}
	
	return syllables
}

// findSyllableEndImproved finds syllable boundaries more accurately
func findSyllableEndImproved(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}
	
	i := start
	hasLeadingVowel := false
	hasVowel := false
	
	// 1. Check for leading vowel
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		hasLeadingVowel = true
		hasVowel = true
		i++
	}
	
	// 2. Get consonant(s)
	consonantStart := i
	consonantCount := 0
	for i < len(runes) && isConsonant(string(runes[i])) {
		consonantCount++
		i++
		
		// Check for valid clusters
		if consonantCount == 2 {
			cluster := string(runes[consonantStart:i])
			if _, isCluster := clusters[cluster]; !isCluster {
				// Not a valid cluster
				if string(runes[consonantStart]) != "ห" {
					// Back up unless it's ห (which can form special clusters)
					i--
					consonantCount--
					break
				}
			}
			break
		}
	}
	
	// 3. Get vowels and tone marks
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) {
			hasVowel = true
			i++
		} else if isToneMark(r) || r == "็" || r == "์" || r == "ํ" || r == "ๆ" {
			i++
		} else {
			break
		}
	}
	
	// 4. Check for final consonant
	if i < len(runes) && isConsonant(string(runes[i])) {
		// Take final consonant if:
		// - We have a vowel
		// - Or it's a closed syllable pattern
		if hasVowel {
			// Check if next position starts a new syllable
			nextIsNewSyllable := false
			if i+1 < len(runes) {
				next := string(runes[i+1])
				nextIsNewSyllable = isLeadingVowel(next) || 
					(isVowel(next) && !hasLeadingVowel) ||
					(isConsonant(next) && hasLeadingVowel)
			}
			
			if !nextIsNewSyllable {
				i++ // Take the final consonant
			}
		} else if consonantCount == 1 && !hasLeadingVowel {
			// CVC pattern with inherent vowel
			i++
		}
	}
	
	// Special check for lone tone marks or special characters
	if i == start && i < len(runes) {
		if isToneMark(string(runes[i])) || string(runes[i]) == "์" {
			i++
		}
	}
	
	return i
}

// findSyllableEnd finds the end of a Thai syllable
func findSyllableEnd(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}
	
	i := start
	hasVowel := false
	hasLeadingVowel := false
	
	// Check for leading vowel (เ แ โ ไ ใ)
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		hasLeadingVowel = true
		hasVowel = true
		i++
	}
	
	// Get initial consonant(s)
	consonantCount := 0
	initialPos := i
	for i < len(runes) && isConsonant(string(runes[i])) {
		i++
		consonantCount++
		// Allow up to 2 consonants (cluster), but be careful with หล patterns
		if consonantCount >= 2 {
			// Check if it's a valid cluster
			clusterStr := string(runes[initialPos:i])
			if _, isCluster := clusters[clusterStr]; !isCluster {
				// Not a valid cluster, back up one
				if consonantCount == 2 && string(runes[initialPos]) != "ห" {
					i--
					consonantCount--
				}
			}
			break
		}
	}
	
	// Get vowels and marks
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) || isToneMark(r) || r == "็" || r == "์" || r == "ํ" || r == "ๆ" {
			i++
			if isVowel(r) {
				hasVowel = true
			}
		} else {
			break
		}
	}
	
	// Check for final consonant
	if i < len(runes) && isConsonant(string(runes[i])) {
		// Special cases for certain finals
		if hasLeadingVowel {
			// With leading vowel, always take final consonant
			i++
		} else if hasVowel {
			// Has vowel, take final if not starting new syllable
			nextIsVowel := i+1 < len(runes) && (isVowel(string(runes[i+1])) || isLeadingVowel(string(runes[i+1])))
			if !nextIsVowel {
				i++
			}
		} else if consonantCount == 1 {
			// Single consonant with no vowel - it's CVC pattern with inherent vowel
			i++
		}
	}
	
	// Ensure we moved forward
	if i == start {
		return start + 1
	}
	
	return i
}

// transliterateSyllable converts a Thai syllable to Paiboon
func transliterateSyllable(syllable string) string {
	// Special cases and common words (including Sanskrit/Pali loanwords)
	specialCases := map[string]string{
		// รร patterns (Sanskrit/Pali double ร)
		"ธรรม": "tam", "กรรม": "gam", "พรรค": "pák", "วรรค": "wák",
		"สรร": "sǎn", "บรร": "ban", "จรร": "jan",
		"รรม": "am", "รรค": "ák", "รรณ": "an", "รรพ": "áp",
		// ทย patterns
		"ทย": "tá~yá", "วิทย": "wít-tá~yá", "วิทยุ": "wít-tá~yú",
		"วิทยา": "wít-tá~yaa", "ศึกษา": "sʉ̀k-sǎa",
		// Common irregular words
		"สัตว์": "sàt", "จริง": "jing", "ทราบ": "sâap",
		"ศิลป": "sǐn-lá~bpà", "ศิลปะ": "sǐn-lá~bpà",
		// Basic common words
		"นอน": "nɔɔn", "แดง": "dɛɛng", "โชค": "chôok", "ลูก": "lûuk",
		"เขียว": "kǐao", "สวัส": "sàwàt", "อร่อ": "àròɔ",
		"สวัสดี": "sàwàtdii", "ขอบ": "kɔ̀ɔp", "คุณ": "kun",
		"ความ": "kwaam", "สุข": "sùk", "อร่อย": "àròɔi",
		"ไม้": "mái", "สวย": "sǔai", "ขอบคุณ": "kɔ̀ɔp-kun",
		"ความสุข": "kwaam-sùk", "ภาษา": "paasǎa", "ภาษาไทย": "paasǎa-tai",
		"ประ": "bprà", "เทศ": "têet", "ประเทศ": "bpràtêet",
		"ประเทศไทย": "bpràtêet-tai",
		// More Sanskrit/Pali loanwords
		"สงฆ์": "sǒng", "นิพพาน": "níp-paan", "ปรินิพพาน": "bpà~rí-níp-paan",
		"ประสงค์": "bprà~sǒng", "มนต์": "mon", "สวดมนต์": "sùuat-mon",
		"อภัย": "à~pai", "เมตตา": "mêet-dtaa", "กรุณา": "gà~rú~naa",
		// Common prefixes/suffixes
		"ระ": "rá", "กระ": "grà", "ตระ": "dtrà",
		// Vowel patterns that are commonly misparsed
		"งอ": "ngɔɔ", "งา": "ngaa", "งู": "nguu",
		"อยู่": "yùu", "อยาก": "yàak",
	}
	
	if trans, ok := specialCases[syllable]; ok {
		return trans
	}

	// Remove silent consonants (consonant + ์) before parsing
	cleanedSyllable := RemoveSilentConsonants(syllable)

	// Parse syllable components
	components := parseSyllableComponents(cleanedSyllable)
	
	// Build transliteration
	result := components.Initial + components.Vowel + components.Final
	
	// Apply tone
	result = applyTone(result, components)
	
	return result
}

// SyllableComponents represents the parts of a Thai syllable
type SyllableComponents struct {
	Initial     string // Initial consonant(s) sound
	Vowel       string // Vowel sound
	Final       string // Final consonant sound
	ToneMark    string // Tone mark if any
	InitialThai string // Original Thai initial for tone class
}

// parseSyllableComponents breaks down a Thai syllable
func parseSyllableComponents(syllable string) SyllableComponents {
	comp := SyllableComponents{}
	runes := []rune(syllable)
	i := 0
	
	leadingVowel := ""
	vowelMarks := ""
	finalCons := ""
	
	// Check for leading vowel
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		leadingVowel = string(runes[i])
		i++
	}
	
	// Special case: อย as initial (as in อยู่)
	if i < len(runes)-1 && string(runes[i]) == "อ" && string(runes[i+1]) == "ย" {
		// Check if followed by vowel (not final position)
		if i+2 < len(runes) && isVowel(string(runes[i+2])) {
			comp.Initial = "y"
			comp.InitialThai = "อ"
			i += 2 // Skip อย
		} else {
			// อย at end would be different
		}
	}
	
	// Get initial consonant(s) if not already set
	if comp.Initial == "" {
		initialCons := ""
		for i < len(runes) && isConsonant(string(runes[i])) {
			initialCons += string(runes[i])
			i++
			
			// Check for หล pattern - ห is silent, ล is the real initial
			if initialCons == "หล" {
				comp.Initial = "l" 
				comp.InitialThai = "ห" // For tone class
				initialCons = "หล"
				break
			}
			
			// Special case: สว cluster in สวัสดี type words
			if initialCons == "ส" && i < len(runes) && string(runes[i]) == "ว" {
				// Check if next is ั (sara a) to form ัว vowel
				if i+1 < len(runes) && string(runes[i+1]) == "ั" {
					// ส + วั pattern
					initialCons = "ส"
					vowelMarks = "วั"
					i += 2 // Skip ว and ั
					break
				}
			}
			if len([]rune(initialCons)) >= 2 {
				break
			}
		}
		
		// Store Thai initial for tone class if not set
		if initialCons != "" && comp.InitialThai == "" {
			comp.InitialThai = string([]rune(initialCons)[0])
		}
		
		// Check for cluster if Initial not set
		if comp.Initial == "" {
			if cluster, ok := clusters[initialCons]; ok {
				comp.Initial = cluster
			} else if initialCons != "" {
				// Use first consonant only
				firstCons := string([]rune(initialCons)[0])
				if trans, ok := initialConsonants[firstCons]; ok {
					comp.Initial = trans
				}
				// If there's a second consonant not in cluster, it might be part of vowel or final
				if len([]rune(initialCons)) > 1 {
					secondCons := string([]rune(initialCons)[1])
					if secondCons == "ว" && vowelMarks == "" {
						vowelMarks = "ว" + vowelMarks
					} else if _, isCluster := clusters[initialCons]; !isCluster {
						// Not a cluster, treat second consonant as final
						finalCons = secondCons
					}
				}
			}
		}
		
		// Special case for อ as initial with ร following
		if initialCons == "อ" && i < len(runes) && string(runes[i]) == "ร" {
			comp.Initial = "" // อ is silent initially before ร
			vowelMarks = "อร"
			i++
		}
	}
	
	// Process remaining characters
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) || r == "็" {
			vowelMarks += r
			i++
		} else if isToneMark(r) {
			comp.ToneMark = r
			i++
		} else if r == "์" {
			// Thanthakhat - silence marker
			i++
		} else if isConsonant(r) {
			// Final consonant
			finalCons = r
			i++
		} else {
			i++
		}
	}
	
	// Determine vowel sound
	comp.Vowel = determineVowelSound(leadingVowel, vowelMarks, finalCons)
	
	// Set final consonant sound
	if finalCons != "" {
		if trans, ok := finalConsonants[finalCons]; ok {
			comp.Final = trans
		}
	}
	
	return comp
}

// determineVowelSound determines the vowel sound from Thai components
func determineVowelSound(leading, marks, final string) string {
	// Handle complex vowel patterns first
	if leading == "เ" {
		if marks == "ีย" || marks == "ียว" {
			return "iia"
		} else if marks == "ือ" {
			return "ʉʉa"  
		} else if marks == "า" {
			return "ao"
		} else if marks == "อ" {
			return "əə"
		} else if marks == "ิ" {
			return "əə"
		} else if marks == "็" {
			return "e"
		} else if marks == "" && final == "ย" {
			// เ-ย pattern as in เลย
			return "əəi"
		} else if marks == "" && final != "" {
			return "ee" // เ-C as in เห็ด
		} else if marks == "" {
			return "ee"
		} else if marks == "ะ" {
			return "e"
		} else if marks == "าะ" {
			return "ɔ"
		}
	} else if leading == "แ" {
		if marks == "็" {
			return "ɛ"
		} else if marks == "ะ" {
			return "ɛ"
		} else if marks == "" && final != "" {
			return "ɛɛ" // แ-C as in แดง
		} else {
			return "ɛɛ"
		}
	} else if leading == "โ" {
		if marks == "ะ" {
			return "o"
		} else if marks == "" && final != "" {
			return "oo" // โ-C as in โชค
		} else {
			return "oo"
		}
	} else if leading == "ไ" || leading == "ใ" {
		return "ai"
	}
	
	// Special patterns with ว
	if marks == "้ว" || marks == "ว" {
		// ด้วย pattern
		if final == "ย" {
			return "uai"
		}
		// Otherwise ว acts as ัว
		return "ua"
	}
	
	// Diphthongs and complex vowels
	if marks == "ียว" {
		return "iao"
	} else if marks == "ือ" {
		if final == "น" {
			// เปื้อน pattern
			return "ʉʉa"
		}
		return "ʉʉ"
	} else if marks == "วั" || marks == "ัว" {
		return "ua"
	} else if marks == "ิว" || marks == "ิ้ว" {
		// นิ้ว pattern
		return "i"
	} else if marks == "อร" {
		return "ɔɔ"
	} else if marks == "อ" && leading == "" {
		return "ɔɔ"
	}
	
	// Simple vowel marks
	switch marks {
	case "ะ":
		return "a"
	case "ั":
		return "a"
	case "า":
		return "aa"
	case "ิ", "ิ้":
		return "i"
	case "ี", "ี้":
		return "ii"
	case "ึ":
		return "ʉ"
	case "ื", "ื้":
		return "ʉʉ"
	case "ุ":
		return "u"
	case "ู", "ู้":
		return "uu"
	case "ำ":
		return "am"
	case "็อ":
		return "ɔ"
	}
	
	// Inherent vowel (no explicit vowel mark)
	if leading == "" && marks == "" {
		if final == "" {
			return "ɔɔ" // Open syllable
		}
		return "o" // Closed syllable with short o
	}
	
	return ""
}

// applyTone applies tone marks to the transliteration
func applyTone(text string, comp SyllableComponents) string {
	// Determine tone class
	toneClass := "mid"
	if highClass[comp.InitialThai] {
		toneClass = "high"
	} else if lowClass[comp.InitialThai] {
		toneClass = "low"
	}
	
	// Determine if syllable is live or dead
	// Live: ends in sonorant (m, n, ng, y, w) or long vowel
	// Dead: ends in stop (p, t, k) or short vowel
	isLive := false
	
	// Check final consonant
	if comp.Final == "" || comp.Final == "n" || comp.Final == "m" || comp.Final == "ng" || comp.Final == "i" || comp.Final == "o" {
		isLive = true
	} else if comp.Final == "p" || comp.Final == "t" || comp.Final == "k" {
		isLive = false
	}
	
	// Check vowel length (long vowels make syllable live)
	longVowels := []string{"aa", "ii", "ʉʉ", "uu", "ee", "ɛɛ", "oo", "ɔɔ", "əə", "iia", "ʉʉa", "uua"}
	for _, lv := range longVowels {
		if strings.Contains(comp.Vowel, lv) {
			isLive = true
			break
		}
	}
	
	// Diphthongs are also live
	if strings.Contains(comp.Vowel, "ai") || strings.Contains(comp.Vowel, "ao") {
		isLive = true
	}
	
	// Get tone number based on Thai tone rules
	toneNum := 0 // mid tone by default
	
	if comp.ToneMark == "" {
		// No tone mark - use inherent tone rules
		switch toneClass {
		case "low":
			if isLive {
				toneNum = 0 // mid tone
			} else {
				toneNum = 2 // high tone (short dead syllable)
			}
		case "mid":
			if isLive {
				toneNum = 0 // mid tone
			} else {
				toneNum = 1 // low tone
			}
		case "high":
			if isLive {
				toneNum = 4 // rising tone
			} else {
				toneNum = 1 // low tone
			}
		}
	} else {
		// Apply tone mark rules
		switch comp.ToneMark {
		case "่": // mai ek
			switch toneClass {
			case "low":
				toneNum = 3 // falling
			case "mid":
				toneNum = 1 // low
			case "high":
				toneNum = 1 // low
			}
		case "้": // mai tho  
			switch toneClass {
			case "low":
				toneNum = 2 // high
			case "mid":
				toneNum = 3 // falling
			case "high":
				toneNum = 3 // falling
			}
		case "๊": // mai tri
			if toneClass == "mid" {
				toneNum = 2 // high
			}
			// Ignored on high/low class
		case "๋": // mai jattawa
			if toneClass == "mid" {
				toneNum = 4 // rising
			}
			// Ignored on high/low class
		}
	}
	
	// Add tone mark to the romanization using proper grapheme handling
	if toneNum == 0 {
		return text // No tone mark for mid tone
	}
	
	marks := map[int]string{
		1: "\u0300", // grave (low)
		2: "\u0301", // acute (high) 
		3: "\u0302", // circumflex (falling)
		4: "\u030C", // caron (rising)
	}
	
	// Find first vowel and add tone mark properly
	graphemes := uniseg.NewGraphemes(text)
	var result strings.Builder
	tonePlaced := false
	
	for graphemes.Next() {
		cluster := graphemes.Str()
		if !tonePlaced && len(cluster) > 0 {
			for _, r := range cluster {
				if isRomanVowel(r) {
					result.WriteString(cluster + marks[toneNum])
					tonePlaced = true
					break
				}
			}
			if !tonePlaced {
				result.WriteString(cluster)
			}
		} else {
			result.WriteString(cluster)
		}
	}
	
	if !tonePlaced {
		return text // No vowel found
	}
	return result.String()
}

// Helper functions

// RemoveSilentConsonants removes consonants followed by ์ (thanthakhat)
// which marks them as silent in Thai orthography.
// Handles both: consonant + ์ and consonant + vowel + ์
// Exported for use by translitkit providers.
func RemoveSilentConsonants(text string) string {
	runes := []rune(text)
	result := make([]rune, 0, len(runes))

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Check if this consonant is followed by ์ (directly)
		if i+1 < len(runes) && runes[i+1] == '์' {
			// Skip both the consonant and the ์
			i++ // Skip the ์ as well (loop will increment again)
			continue
		}

		// Check if this consonant + vowel is followed by ์
		// Pattern: consonant + vowel + ์ (e.g., ธุ์ in พันธุ์)
		if i+2 < len(runes) && isConsonantRune(r) && isVowelRune(runes[i+1]) && runes[i+2] == '์' {
			// Skip consonant, vowel, and ์
			i += 2 // Skip vowel and ์ (loop will increment again)
			continue
		}

		result = append(result, r)
	}

	return string(result)
}

func isConsonantRune(r rune) bool {
	return strings.ContainsRune("กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรฤลฦวศษสหฬอฮ", r)
}

func isVowelRune(r rune) bool {
	return strings.ContainsRune("ะัาิีึืุูเแโใไๅำ", r)
}

func isConsonant(s string) bool {
	return strings.Contains("กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรฤลฦวศษสหฬอฮ", s)
}

func isVowel(s string) bool {
	return strings.Contains("ะัาิีึืุูเแโใไๅำ", s)
}

func isLeadingVowel(s string) bool {
	return s == "เ" || s == "แ" || s == "โ" || s == "ไ" || s == "ใ"
}

func isToneMark(s string) bool {
	return strings.Contains("่้๊๋", s)
}

func isRomanVowel(r rune) bool {
	return strings.ContainsRune("aeiouəɛɔʉ", r)
}

// Testing functions
func test(th, trg string) {
	r := TransliterateWordRulesOnly(th)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	r, _, _ = transform.String(t, r)
	fmt.Println(isPassed(r, trg), th, "\t\t\t\t", r, "→", trg)
}

func isPassed(result, target string) string {
	str := "FAIL"
	c := color.FgRed.Render
	if result == target {
		c = color.FgGreen.Render
		str = "OK"
	}
	return fmt.Sprintf(c(str))
}

// Data loading
var words []string
var m = make(map[string]string)
var re = regexp.MustCompile(`(.*),(.*\p{Thai}.*)`)

func init() {
	// Use embedded filesystem for vocab files
	entries, err := fs.ReadDir(vocabFS, "csv")
	check(err)

	for _, e := range entries {
		dat, err := fs.ReadFile(vocabFS, "csv/"+e.Name())
		check(err)
		arr := strings.Split(string(dat), "\n")

		for _, str := range arr {
			raw := re.FindStringSubmatch(str)
			if len(raw) == 0 {
				continue
			}
			row := strings.Split(raw[2], ",")[:2]
			th := html.UnescapeString(row[0])
			translit := html.UnescapeString(row[1])

			// Add to test data
			words = append(words, th)
			m[th] = translit

			// Build dictionary
			dictionary[th] = translit

			// Try to extract single syllables for syllable dictionary
			// Add short words and very common syllables
			if !strings.Contains(th, " ") {
				if len([]rune(th)) <= 5 && !strings.Contains(translit, "-") {
					syllableDict[th] = translit
				} else if len([]rune(th)) <= 3 {
					// Very short words are almost always single syllables
					syllableDict[th] = translit
				}
			}
		}
	}

	// Extract syllables from multi-syllable dictionary entries
	extractSyllablesFromDictionary()

	fmt.Printf("Dictionary built: %d entries, %d syllables\n", len(dictionary), len(syllableDict))
}

// extractSyllablesFromDictionary extracts individual syllables from multi-syllable
// dictionary entries to expand the syllable dictionary for maximal matching
func extractSyllablesFromDictionary() {
	// CRITICAL: Sort dictionary keys for deterministic iteration order
	// Otherwise if two words share a syllable with different romanizations,
	// whichever is processed first wins, causing entropy in measured accuracy
	sortedKeys := make([]string, 0, len(dictionary))
	for k := range dictionary {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Process entries with hyphens (multi-syllable words)
	for _, th := range sortedKeys {
		translit := dictionary[th]
		if strings.Contains(translit, "-") {
			// Split Thai text into syllables using rule-based extraction
			thaiSyllables := ExtractSyllables(th)
			// Split romanization by hyphens
			romanSyllables := strings.Split(translit, "-")

			// Only use if counts match (reliable mapping)
			if len(thaiSyllables) == len(romanSyllables) {
				for i, thaiSyl := range thaiSyllables {
					romanSyl := romanSyllables[i]
					// Only add if not already in dictionary and reasonable length
					if _, exists := syllableDict[thaiSyl]; !exists {
						if len([]rune(thaiSyl)) >= 2 && len([]rune(thaiSyl)) <= 6 {
							syllableDict[thaiSyl] = romanSyl
						}
					}
				}
			}
		}
	}

	// Also add common Thai syllable patterns from special cases
	for th, translit := range specialCasesGlobal {
		if !strings.Contains(translit, "-") && len([]rune(th)) <= 5 {
			if _, exists := syllableDict[th]; !exists {
				syllableDict[th] = translit
			}
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
func main() {
	// Define command line flags
	dictionaryCheck := flag.Bool("dictionary-check", false, "Test transliterator accuracy against dictionary")
	analyzeFlag := flag.Bool("analyze", false, "Analyze specific failure cases")
	testFile := flag.Bool("test", false, "Test transliterator on test.txt")
	noDictionary := flag.Bool("no-dict", false, "Disable dictionary lookup for testing pure rules")
	flag.Parse()
	
	// Clean up pythainlp on exit
	defer func() {
		if Manager.nlpManager != nil {
			ctx := context.Background()
			Manager.nlpManager.Stop(ctx)
		}
	}()
	
	// Handle flags
	if *dictionaryCheck {
		testDictionary(!*noDictionary) // Pass true to use dictionary, false to use pure rules
		return
	}
	
	if *analyzeFlag {
		analyzeFailures()
		return
	}
	
	if *testFile {
		testTransliterate()
		return
	}
	
	// Default test
	fmt.Println("Use -dictionary-check, -analyze, or -test flags to run tests")
}*/

func testWiktionary() {
	test("น้ำ", "nám")
	test("ธรรม", "tam")
	test("บาด", "bàat")
	test("บ้า", "bâa")
	test("แข็ง", "kɛ̌ng")
	test("แกะ", "gɛ̀")
	test("แดง", "dɛɛng")
	test("เกาะ", "gɔ̀")
	test("นอน", "nɔɔn")
	test("พ่อ", "pɔ̂ɔ")
	test("เห็ด", "hèt")
	test("เตะ", "dtè")
	test("เยอะ", "yə́")
	test("เดิน", "dəən")
	test("ตก", "dtòk")
	test("โต๊ะ", "dtó")
	test("โชค", "chôok")
	test("คิด", "kít")
	test("อีก", "ìik")
	test("จี้", "jîi")
	test("ลึก", "lʉ́k")
	test("รึ", "rʉ́")
	test("ชื่อ", "chʉ̂ʉ")
	test("คุก", "kúk")
	test("ลูก", "lûuk")
	test("ปู", "bpuu")
	test("เตียง", "dtiiang")
	test("เมีย", "miia")
	test("เรือ", "rʉʉa")
	test("นวด", "nûuat")
	test("ตัว", "dtuua")
	test("ไม่", "mâi")
	test("ใส่", "sài")
	test("วัย", "wai")
	test("ไทย", "tai")
	test("ไม้", "mái")
	test("หาย", "hǎai")
	test("ซอย", "sɔɔi")
	test("เลย", "ləəi")
	test("โดย", "dooi")
	test("ทุย", "tui")
	test("สวย", "sǔai")
	test("เรา", "rao")
	test("ขาว", "kǎao")
	test("แมว", "mɛɛo")
	test("เร็ว", "reo")
	test("หิว", "hǐu")
	test("เขียว", "kǐao")
	test("ทำ", "tam")
}
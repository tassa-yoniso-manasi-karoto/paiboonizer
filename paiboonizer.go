package main

import (
	"strings"
	"unicode/utf8"
	"regexp"
	"fmt"
	//"log"
	//"os/exec"
	"os"
	"github.com/gookit/color"
	"path/filepath"
	"html"
	"slices"
	
	"unicode"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"github.com/k0kubun/pp"
)

const (
	LongVwl = iota
	ShortVwl
	Vwl
	Cons
	Other
)

var (
	lows = []string{"ค", "ฅ", "ฆ", "ง", "ช", "ซ", "ฌ", "ญ", "ฑ", "ฒ", "ณ", "ท", "ธ", "น", "พ", "ฟ", "ภ", "ม", "ย", "ร", "ล", "ว", "ฬ", "ฮ"}
	mids = []string{"ก", "จ", "ฎ", "ฏ", "ด", "ต", "บ", "ป", "อ"}
	highs = []string{"ข", "ฃ", "ฉ", "ฐ", "ถ", "ผ", "ฝ", "ศ", "ษ", "ส", "ห"}
	
)

type UnitType struct {
	Before, MainPart string
	Translit string
	Type    int
	AddedAt string
}

type SyllableType struct {
	Units                      []UnitType
	Alive                      bool
	HasToneMark, HasLongVowel  bool
	InitialConsonantClass      int
}


func ThaiToRoman(str string) (out string) {
	var (
		consonants           = slices.Concat(lows, mids, highs)
		Befores              []string
		MainVowelParts       []string
		
		//cons, vowel, closing   UnitType
		Units             []UnitType
		
		Syllables            []SyllableType
		before, mainpart, char string
	)
	for _, vowel := range vowels {
		Befores        = append(Befores, vowel.Before)
		MainVowelParts = append(MainVowelParts, vowel.MainPart)
	}
	orig := str
	str = strings.Trim(str, "่้็๋์")
	f := func(loc string) {
		unit := UnitType{
			Before: before,
			MainPart: mainpart,
			Type: Vwl,
			AddedAt: loc,
		}
		unit.Translit = unit.GetTranslit(str)
		if mainpart != "" {
			Units = append(Units, unit)
			//Syllable.Units = append(Syllable.Units, unit)
		}
		mainpart, before = "", ""
	}
	for str != "" {
		r, _ := utf8.DecodeRuneInString(str)
		char = string(r)
		for _, MainVowelPart := range MainVowelParts {
			if MainVowelPart != "" && strings.HasPrefix(str, MainVowelPart) {
				mainpart = MainVowelPart
				f("1")
				char = MainVowelPart
				break
			}
		}
		if contains(consonants, char) {
			unit := UnitType{
				MainPart: char,
				Type: Cons,
				AddedAt: "2",
			}
			unit.Translit = unit.GetTranslit(str)
			Units = append(Units, unit)
		} else if contains(Befores, char) {
			if before != "" {
				f("3")
			}
			before = char
		}
		str = strings.TrimPrefix(str, char)
	}	
	for i, Unit := range Units {
		if Unit.Type == Vwl {
			var Syllable SyllableType
			/*defer func() {
				if x := recover(); x != nil {
					println(orig)
					fmt.Println(i, Unit.Translit, Unit.Type)
					pp.Println(Units)
					panic("")
				}
			}()*/
			Syllable.Units = append(Syllable.Units, Units[i-1])
			Syllable.Units = append(Syllable.Units, Unit)
			Syllables = append(Syllables, Syllable)
		} else {
			Syllables = append(Syllables, SyllableType{Units:[]UnitType{Unit}})
		}
	}
	if orig == "เป็นกังวล" {
		pp.Println(Units)
		pp.Println(Syllables)
		os.Exit(0)
	}
	/*out += cons.Translit + vowel.GetVowelTranslit(str) + closing.Translit*/
	return
} 

/*func (Syllable *SyllableType) String() (s string) {
	for _, Unit := range Syllable.Units {
		
	}
	return
}
*/

func (unit UnitType) GetTranslit(str string) (translit string) {
	for _, ref := range vowels {
		if ref.Before == unit.Before && ref.MainPart == unit.MainPart {
			translit = ref.Translit
		}
	}
	if translit == "" {
		translit = scheme[unit.MainPart]
	}
	/*if unit.Before == "" && unit.MainPart == "" && len(str) > 1 {
		translit = "o"
	}*/
	return
}














func test(th, trg string) {
	r := ThaiToRoman(th)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	r, _, _ = transform.String(t, r)
	//trg, _, _ = transform.String(t, trg)
	fmt.Println(isPassed(r, trg), th, "\t\t\t\t", r, "→", trg)
}

func isPassed(th, trg string) string {
	str := "FAIL"
	c := color.FgRed.Render
	if th == trg {
		c = color.FgGreen.Render
		str = "OK"
	}
	return fmt.Sprintf(c(str))
}


var words []string
var m = make(map[string]string)
var re = regexp.MustCompile(`(.*),(.*\p{Thai}.*)`)

func init() {
	pp.Println("Init")
	path := "/home/voiduser/go/src/Thaidict/manual vocab/"
	a, err := os.ReadDir(path)
	check(err)
	for _, e := range a {
		file := filepath.Join(path, e.Name())
		dat, err := os.ReadFile(file)
		check(err)
		arr := strings.Split(string(dat), "\n")
		//out := ""
		//fmt.Printf("%#v\n", arr)
		for _, str := range arr {
			raw := re.FindStringSubmatch(str)
			if len(raw) == 0 {
				//fmt.Println(str)
				continue
			}
			row := strings.Split(raw[2], ",")[:2]
			//fmt.Printf("%#v\n", row)
			/*if len(lineItems) < 7 {
				if str != "" {
					fmt.Println("inf to 7:", str)
				}
				continue
			}*/
			th := html.UnescapeString(row[0])
			translit := html.UnescapeString(row[1])
			words = append(words, th)
			m[th] = translit
			//out += th + "\t" + translit + "\n"
		}
	}
	//_ = os.WriteFile("/home/voiduser/go/src/Thaidict/translitdb.tsv", []byte(out), 0644)
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func contains[T comparable](arr []T, i T) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}



func main() {
	fmt.Println(len(words))
	for _, th := range words {
		test(th, m[th])
	}


	//s := "ถามเขาหน่อยว่า มีมือถือหรือเปล่า"
	//s := "อย่าวิจารณ์คนอื่นในที่สาธารณะ"
	//fmt.Printf("%v\n", m)
	/*binary, lookErr := exec.LookPath("thainlp")
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{"thainlp", "tokenize", "word", s}
	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}*/
	
	/*out, err := exec.Command("thainlp", "tokenize", "subword", s).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
	fmt.Println(ThaiToRoman(string(out), 1))*/
	//fmt.Println("yaa-wi-jaan-kon-ʉʉn-nai-tii-saa-taa-ra~na")
	
	
	
	
	
	/*
	test("เฮิ็้ย", "hə́i")
	test("เฉิ็ย", "chə̌i")
	test("เงิ็น", "ngən")
	test("เกดส", "gèets") //เกตส์ (Gates)
	test("มันส", "mans") //มันส์
	test("ไอ๊ส", "áis") //ไอซ์ (ice)
	test("เซ็กส", "séks") //เซ็กส์ (sex)
	test("เอ็๊กส", "éks") //เอกซ์ & เอ็กซ์ & เอ๊กซ์ (ex)
	test("เฮ้าส", "háos") //เฮาส์ & เฮ้าส์ (house)
	test("เม้าส", "máos") //เมาส์ & เม้าส์ (mouse)
	test("ทฺรำ-เป็ด", "tram-bpèt") //ทรัมเป็ต
	test("ห็อย", "hɔ̌i")
	test("หฺม็อย", "mɔ̌i")
	test("หฺมั่น-โถว", "màn-tǒow")
	test("เด๊ด-สะ-มอ-เร่", "déet-sà-mɔɔ-rêe")
	test("เห", "hěe")
	test("แคฺล", "klɛɛ")
	test("แคล", "kɛɛl")
	test("เพฺล", "plee")
	test("เพล", "peel")
	test("เปฺล", "bplee")
	test("เปล", "bpeel")
	test("เบล", "beel")
	test("เซล", "seel")
	test("โพล", "pool")
	test("รา-ชา-ทิ-ราด", "raa-chaa-tí-râat")
	test("ขฺวน-ขฺวาย", "kwǒn-kwǎai") //ขวนขวาย Only the word ขวน read as kwǒn instead of kǔuan.
	test("ข่วน", "kùuan")
	test("หอน", "hɔ̌ɔn")
	test("โหน", "hǒon") // ห้อยโหน homograph issue
	test("สะ-โหฺน", "sà-nǒo") // โสน homograph issue
	test("แหน", "hɛ̌ɛn") // หวงแหน homograph issue
	test("แหฺน", "nɛ̌ɛ") // จอกแหน homograph issue
	test("แถ็ว", "tɛ̌o") // แถว
	test("ซวง", "suuang")
	test("น้ำ", "nám")
	test("หฺมาย", "mǎai")
	test("แห็่ง", "hɛ̀ng")
	test("หน", "hǒn")
	test("เหด-สุด-วิ-ไส", "hèet-sùt-wí-sǎi")
	test("ไหฺย่", "yài")
	test("หก", "hòk")
	test("หอย", "hɔ̌ɔi")
	test("กับ", "gàp")
	test("ธรรม", "tam")
	test("ปฺระ-ชา", "bprà-chaa")
	test("นะ-คอน", "ná-kɔɔn")
	test("บาด", "bàat")
	test("บ้า", "bâa")
	test("แข็ง", "kɛ̌ng")
	test("แกะ", "gɛ̀")
	test("แดง", "dɛɛng")
	test("แปฺล", "bplɛɛ")
	test("ผฺล็อง", "plɔ̌ng")
	test("เกาะ", "gɔ̀")
	test("นอน", "nɔɔn")
	test("พ่อ", "pɔ̂ɔ")
	test("เห็ด", "hèt")
	test("เล็่น", "lên")
	test("เตะ", "dtè")
	test("เพฺลง", "pleeng")
	test("เท-วี", "tee-wii")
	test("เยอะ", "yə́")
	test("เดิน", "dəən")
	test("เผฺลอ", "plə̌ə")
	test("ตก", "dtòk")
	test("โต๊ะ", "dtó")
	test("โชค", "chôok")
	test("โม-โห", "moo-hǒo")
	test("คิด", "kít")
	test("มิ-ถุน", "mí-tǔn")
	test("หิ-มะ", "hì-má")
	test("อีก", "ìik")
	test("จี้", "jîi")
	test("ลึก", "lʉ́k")
	test("รึ", "rʉ́")
	test("กฺลืน", "glʉʉn")
	test("ชื่อ", "chʉ̂ʉ")
	test("คุก", "kúk")
	test("จุ-ฬา", "jù-laa")
	test("ลูก", "lûuk")
	test("ปู", "bpuu")
	test("เดี๊ยะ", "día")
	test("เปาะ-เปี๊ยะ", "bpɔ̀-bpía")
	test("ปอ-เปี๊ยะ", "bpɔɔ-bpía")
	test("เปฺรี๊ยะ", "bpría")
	test("เตียง", "dtiiang")
	test("เมีย", "miia")
	test("เอือะ", "ʉ̀a")
	test("เรื่อง", "rʉ̂ʉang")
	test("เรือ", "rʉʉa")
	test("ผฺลัวะ", "plùa")
	test("นวด", "nûuat")
	test("ตัว", "dtuua")
	test("ไม่", "mâi")
	test("ใส่", "sài")
	test("วัย", "wai")
	test("ไทย", "tai")
	test("ไม้", "mái")
	test("หาย", "hǎai")
	test("ผฺล็อย", "plɔ̌i")
	test("ซอย", "sɔɔi")
	test("เลย", "ləəi")
	test("โดย", "dooi")
	test("ทุย", "tui")
	test("เหฺนื่อย", "nʉ̀ai")
	test("สวย", "sǔai")
	test("เรา", "rao")
	test("ขาว", "kǎao")
	test("แมว", "mɛɛo")
	test("เกอว", "gəəo")
	test("เร็ว", "reo")
	test("เอว", "eeo")
	test("หิว", "hǐu")
	test("เขียว", "kǐao")
	test("ทำ", "tam")*/

}


var scheme = map[string]string{
	"ภ": "p",
	"ม": "m",
	"ล": "l",
	"พ": "p",
	"ก": "g",
	"ข": "k",
	"ค": "k",
	"ฆ": "k",
	"ง": "ng",
	"จ": "j",
	"ฉ": "ch",
	"ช": "ch",
	"ฌ": "ch",
	"ญ": "y",
	"ฏ": "dt",
	"ฐ": "t",
	"ฑ": "t",
	"ฒ": "t",
	"ณ": "n",
	"ต": "dt",
	"ถ": "t",
	"ท": "t",
	"ธ": "t",
	"น": "n",
	"ป": "bp",
	"ผ": "p",
	"ย": "y",
	"ร": "r",
	"ว": "w", // /!\ 
	"ส": "s",
	"ห": "h",
	"ฬ": "l",
	//"อ": "ɔɔ",
	// ɔɔ̂ɔ̌ɔ̀ɔ́ ɛɛ̂ɛ̌ɛ̀ɛ́ ɜɜ̂ɜ̌ɜ̀ɜ́ ǎǒǔǐ ï ᵐ
	
	"๐": "0",
	"๑": "1",
	"๒": "2",
	"๓": "3",
	"๔": "4",
	"๕": "5",
	"๖": "6",
	"๗": "7",
	"๘": "8",
	"๙": "9",
}

var schemeClosing = map[string]string{ // PAS DE PAIBOON
	"ภ": "p",
	"ม": "m",
	"ล": "n", //←
	"พ": "p",
	"ก": "g",
	"ข": "k",
	"ค": "k",
	"ฆ": "k",
	"ง": "ng",
	"จ": "t", //←
	"ฉ": "", //←
	"ช": "t", //←
	"ฌ": "t", //←
	"ญ": "b", //←
	"ฏ": "t",//←
	"ฐ": "t",
	"ฑ": "t",
	"ฒ": "t",
	"ณ": "n",
	"ต": "t", //←
	"ถ": "t",
	"ท": "t",
	"ธ": "t",
	"น": "n",
	"ป": "p",//←
	"ผ": "",//←
	"ย": "n",//←
	"ร": "n",//←
	"ว": "",//←
	"ส": "t",//←
	"ห": "",//←
	"ฬ": "n",//←
}

//#############################################################################################################

	
var vowels = []UnitType{
	UnitType{ //  	sara e
		Before: "เ",
		MainPart:  "็",
		Translit:   "e",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "าว",
		Translit:   "aao",
	},
	UnitType{
		Before: "",
		MainPart:  "ะ",
		Translit:   "a",
	},
	UnitType{
		Before: "",
		MainPart:  "ั",
		Translit:   "a",
	},
	UnitType{
		Before: "",
		MainPart:  "า",
		Translit:   "aa",
	},
	UnitType{
		Before: "เ",
		MainPart:  "",
		Translit:   "ee",
	},
	UnitType{
		Before: "โ",
		MainPart:  "",
		Translit:   "oo",
	},
	UnitType{
		Before: "",
		MainPart:  "ุ",
		Translit:   "u",
	},
	UnitType{
		Before: "",
		MainPart:  "ู",
		Translit:   "uu",
	},
	UnitType{
		Before:  "",
		MainPart:   "ิ",
		Translit:   "i",
	},
	UnitType{
		Before: "",
		MainPart:  "ี",
		Translit:   "ii",
	},
	UnitType{
		Before: "",
		MainPart:  "ึ",
		Translit:   "ʉ",
	},
	UnitType{
		Before: "",
		MainPart:  "ื",
		Translit:   "ʉʉ",
	},
	UnitType{
		Before: "",
		MainPart:  "ำ",
		Translit:   "am",
	},
	UnitType{
		Before: "แ",
		MainPart:  "",
		Translit:   "ɛɛ",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
	UnitType{
		Before: "",
		MainPart:  "",
		Translit:   "",
	},
}


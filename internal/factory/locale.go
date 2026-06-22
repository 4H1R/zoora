package factory

import (
	"fmt"
	"os"
	"strings"
)

// Locale selects the language of generated text content.
type Locale string

const (
	LocaleEn Locale = "en"
	LocaleFa Locale = "fa"
)

// SeedLangEnv is the env var the factory reads to choose its locale. The seeder
// writes the resolved choice here; defaulting to English keeps existing callers
// (unit tests, dev helpers) on their original output. Reading from config (env)
// on demand avoids package-level mutable global state.
const SeedLangEnv = "SEED_LANG"

// IsFa reports whether the configured locale is Persian.
func IsFa() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(SeedLangEnv)), string(LocaleFa))
}

// T returns fa when the current locale is Persian, otherwise en. Useful for
// callers (the seeder) that build their own literal strings.
func T(en, fa string) string {
	if IsFa() {
		return fa
	}
	return en
}

// Persian word pools used to build deterministic Persian text via the shared
// seeded faker (gofakeit, seed 42) so output stays reproducible across runs.
var (
	faFirstNames = []string{
		"علی", "محمد", "رضا", "حسین", "مهدی", "امیر", "سعید", "حمید", "بهرام", "آرش",
		"سارا", "زهرا", "فاطمه", "مریم", "نگار", "الهام", "پریسا", "شیرین", "نیلوفر", "مینا",
	}
	faLastNames = []string{
		"محمدی", "حسینی", "رضایی", "کریمی", "موسوی", "احمدی", "اکبری", "نجفی", "قاسمی", "صادقی",
		"رحیمی", "جعفری", "کاظمی", "علیزاده", "یوسفی", "شریفی", "نوری", "بهرامی", "سلطانی", "مرادی",
	}
	faCities = []string{
		"تهران", "اصفهان", "شیراز", "مشهد", "تبریز", "یزد", "کرمان", "اهواز", "رشت", "قم",
	}
	faSubjects = []string{
		"ریاضی", "فیزیک", "شیمی", "زیست‌شناسی", "ادبیات", "تاریخ", "جغرافیا", "هنر", "موسیقی",
		"برنامه‌نویسی", "فلسفه", "اقتصاد", "حقوق", "زبان انگلیسی", "علوم رایانه",
	}
	faJobTitles = []string{
		"مدیر", "سرپرست", "هماهنگ‌کننده", "ناظر", "کارشناس", "مشاور", "دستیار", "سرگروه",
	}
	faQuestionStems = []string{
		"کدام گزینه صحیح است",
		"تعریف درست را انتخاب کنید",
		"حاصل عبارت زیر چیست",
		"کدام مورد نادرست است",
		"بهترین پاسخ کدام است",
		"علت اصلی چیست",
	}
	faWords = []string{
		"دانش", "آموزش", "یادگیری", "کلاس", "درس", "تمرین", "پروژه", "آزمون", "نمره", "جلسه",
		"معلم", "دانش‌آموز", "کتاب", "مطالعه", "پاسخ", "سوال", "موضوع", "مفهوم", "نتیجه", "تحلیل",
	}
)

func faPick(pool []string) string { return fake.RandomString(pool) }

// fakeName returns a person's full name in the current locale.
func fakeName() string {
	if IsFa() {
		return faPick(faFirstNames) + " " + faPick(faLastNames)
	}
	return fake.Name()
}

// fakeSentence returns a sentence of roughly n words in the current locale.
func fakeSentence(n int) string {
	if IsFa() {
		if n < 1 {
			n = 1
		}
		words := make([]string, n)
		for i := range words {
			words[i] = faPick(faWords)
		}
		return strings.Join(words, " ") + "."
	}
	return fake.Sentence(n)
}

// Entity-name helpers keep the English templates identical to the originals
// and provide a natural Persian equivalent.

func fakeOrgName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("دانشگاه %s %d", faPick(faCities), id)
	}
	return fmt.Sprintf("%s University %d", fake.Company(), id)
}

func fakeClassName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("کلاس %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s %d", fake.School(), id)
}

func fakeSessionName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("جلسه %d", id)
	}
	return fmt.Sprintf("Session %d", id)
}

func fakeChatName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("گفتگوی %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s Chat %d", fake.Noun(), id)
}

func fakePollName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("نظرسنجی %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s Poll %d", fake.Noun(), id)
}

func fakeQuizTitle(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("آزمون %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s Quiz %d", fake.Noun(), id)
}

func fakeQuestionBankName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("بانک سوال %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s Question Bank %d", fake.Noun(), id)
}

func fakeQuestionText(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("%s؟ (%d)", faPick(faQuestionStems), id)
	}
	return fmt.Sprintf("%s? (%d)", fake.Question(), id)
}

func fakeRoleName(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("نقش %s %d", faPick(faJobTitles), id)
	}
	return fmt.Sprintf("%s Role %d", fake.JobTitle(), id)
}

func fakeGradebookColumnTitle(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("ستون %d", id)
	}
	return fmt.Sprintf("Column %d", id)
}

func fakeOfflineRoomTitle(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("ضبط %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s Recording %d", fake.Noun(), id)
}

func fakePracticeTitle(id uint64) string {
	if IsFa() {
		return fmt.Sprintf("تمرین %s %d", faPick(faSubjects), id)
	}
	return fmt.Sprintf("%s Practice %d", fake.Noun(), id)
}

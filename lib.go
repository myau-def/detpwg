package detpwg

import (
	"crypto/hkdf"
	"crypto/sha256"
	"errors"
	"math/big"
	"slices"

	"golang.org/x/crypto/argon2"
)

// GenerateMasterKey генерирует мастер ключ из мастер имени и мастер пароля с помощью Argon2id
func GenerateMasterKey(master, password string) ([]byte, error) {
	// Argon2id предпочитает соль из 16 байт, поэтому растягиваем мастер имя до 16 уникальных байт
	salt, err := hkdf.Key(sha256.New, []byte(master), []byte("deterministic password generation"), "master key salt expand", 16)
	if err != nil {
		return nil, err
	}
	return argon2.IDKey([]byte(password), salt, 3, 256*1024, 4, 32), nil
}

// GeneratePasswordEntropy генерирует из логина и мастер ключа энтропию для конечного пароля
func GeneratePasswordEntropy(key *[]byte, login *Login) ([]byte, error) {
	return hkdf.Key(sha256.New, *key, login.Salt(), "password entropy generation", 128)
}

// RenderPassword создаёт пароль в виде строки заданной длины из энтропии, используя переданный алфавит
func RenderPassword(entropy *[]byte, alphabet Alphabet, length int) (string, error) {
	// Пароль должен содержать по символу из каждого набора,
	// он не может иметь длину меньше, чем требуемое количество символов в нём
	if length < len(alphabet) {
		return "", errors.New("password length must be greater than or equal to count of charsets")
	}

	// Валидируем алфавит + получаем общий пул символов
	poolSet, err := alphabet.ValidateAndPool()
	if err != nil {
		return "", err
	}
	// Нормализуем: сортируем пул для детерминированного положения символов
	pool := poolSet.Pool()

	// Создаём глубокую копию, т.к. alphabet будет мутировать
	alphabet = alphabet.Clone()
	// Инициализируем энтропию как BigInt
	intEntropy := new(big.Int).SetBytes(*entropy)

	// Инициализируем пароль
	password := make([]rune, 0, length)

	// Инициализируем переменные
	base := big.NewInt(int64(len(pool)))
	index := new(big.Int)
	// Продолжаем пока не останется мест только на обязательные символы.
	for length-len(password) != len(alphabet) {
		intEntropy.DivMod(intEntropy, base, index)
		r := pool[index.Uint64()]
		password = append(password, r)

		for i, cs := range alphabet {
			if _, exists := cs.RuneSet[r]; exists {
				alphabet, _ = SwapAndPop(alphabet, i)
				break
			}
		}
	}

	// Вставляем символы из гарантированных наборов не поучавствовавших в выборке
	if len(alphabet) != 0 {
		// Нормализуем
		sets := make([]string, 0, len(alphabet))
		for _, cs := range alphabet {
			sets = append(sets, cs.String())
		}
		slices.Sort(sets)

		// Вставляем символы
		var set string
		for len(sets) != 0 {
			// Выбираем случайный набор
			base.SetInt64(int64(len(sets)))
			intEntropy.DivMod(intEntropy, base, index)
			sets, set = SwapAndPop(sets, int(index.Uint64()))
			pool := []rune(set)

			// Выбираем случайный символ
			base.SetInt64(int64(len(pool)))
			intEntropy.DivMod(intEntropy, base, index)
			r := pool[index.Uint64()]

			// Вставляем символ в пароль
			base.SetInt64(int64(len(password) + 1))
			intEntropy.DivMod(intEntropy, base, index)
			password = SwapToBack(password, int(index.Uint64()), r)
		}
	}

	// if intEntropy < (1 << 128)
	if intEntropy.Cmp(new(big.Int).Lsh(big.NewInt(1), 128)) < 0 {
		return "", errors.New("entropy is over")
	}

	return string(password), nil
}

// GeneratePassword генерирует конечный пароль
func GeneratePassword(key *[]byte, login *Login, alphabet Alphabet, length int) (string, error) {
	entropy, err := GeneratePasswordEntropy(key, login)
	if err != nil {
		return "", err
	}
	return RenderPassword(&entropy, alphabet, length)
}

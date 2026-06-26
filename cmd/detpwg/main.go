package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/myau-def/detpwg"
	"github.com/spf13/cobra"
)

// Переменные для флагов CLI
var (
	info             string
	counter          int
	length           int
	exclude          string
	noLowercase      bool
	includeLowercase bool
	noUppercase      bool
	includeUppercase bool
	noDigits         bool
	includeDigits    bool
	noSpecials       bool
	includeSpecials  bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "detpwg",
		Short: "Детерминированный генератор паролей и ключей",
	}
	
	// ----------------------------------------
	// Команда: fpw <MASTER> <PASSWORD> <SERVICE> [LOGIN]
	// ----------------------------------------
	var fpwCmd = &cobra.Command{
		Use:   "fpw <MASTER> <PASSWORD> <SERVICE> [LOGIN]",
		Short: "Генерация пароля по мастер-имени и мастер-паролю",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			master := args[0]
			password := args[1]
			service := args[2]
			loginName := ""
			if len(args) > 3 {
				loginName = args[3]
			}

			// 1. Инициализируем Мастер-Ключ
			mk, err := detpwg.NewMasterKey(master, password)
			if err != nil {
				return fmt.Errorf("ошибка генерации мастер-ключа: %w", err)
			}

			// 2. Настраиваем логин/аккаунт
			login := detpwg.Login{
				Service: service,
				Login:   loginName,
				Info:    info,
				Counter: counter,
			}

			// 3. Строим алфавит на основе флагов
			alphabet, err := buildAlphabetFromFlags()
			if err != nil {
				return fmt.Errorf("ошибка построения алфавита: %w", err)
			}

			// 4. Генерируем пароль
			pwd, err := mk.GeneratePassword(&login, alphabet, length)
			if err != nil {
				return fmt.Errorf("ошибка рендеринга пароля: %w", err)
			}

			fmt.Println(pwd)
			return nil
		},
	}

	// ----------------------------------------
	// Команда: fkey <KEY_HEX> <SERVICE> [LOGIN]
	// ----------------------------------------
	var fkeyCmd = &cobra.Command{
		Use:   "fkey <KEY_HEX> <SERVICE> [LOGIN]",
		Short: "Генерация пароля по готовому ключу в формате HEX",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyHex := args[0]
			service := args[1]
			loginName := ""
			if len(args) > 2 {
				loginName = args[2]
			}

			// 1. Декодируем мастер-ключ из HEX
			keyBytes, err := hex.DecodeString(keyHex)
			if err != nil {
				return fmt.Errorf("неверный формат HEX-ключа: %w", err)
			}
			mk := detpwg.MasterKey{Key: keyBytes}

			// 2. Настраиваем логин/аккаунт
			login := detpwg.Login{
				Service: service,
				Login:   loginName,
				Info:    info,
				Counter: counter,
			}

			// 3. Строим алфавит
			alphabet, err := buildAlphabetFromFlags()
			if err != nil {
				return fmt.Errorf("ошибка построения алфавита: %w", err)
			}

			// 4. Генерируем пароль
			pwd, err := mk.GeneratePassword(&login, alphabet, length)
			if err != nil {
				return fmt.Errorf("ошибка рендеринга пароля: %w", err)
			}

			fmt.Println(pwd)
			return nil
		},
	}

	// ----------------------------------------
	// Команда: key <MASTER> <PASSWORD>
	// ----------------------------------------
	var keyCmd = &cobra.Command{
		Use:   "key <MASTER> <PASSWORD>",
		Short: "Генерация и вывод мастер-ключа в формате HEX",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			master := args[0]
			password := args[1]
			key, err := detpwg.GenerateMasterKey(master, password)
			if err != nil {
				return fmt.Errorf("ошибка генерации мастер-ключа: %w", err)
			}

			// Выводим ключ в HEX
			fmt.Println(hex.EncodeToString(key))
			return nil
		},
	}

	// Настройка общих флагов генерации для fpw и fkey
	for _, cmd := range []*cobra.Command{fpwCmd, fkeyCmd} {
		cmd.Flags().StringVarP(&info, "info", "i", "account on service", "Информация об аккаунте (часть соли)")
		cmd.Flags().IntVarP(&counter, "counter", "c", 1, "Счетчик генераций (часть соли)")
		cmd.Flags().IntVarP(&length, "length", "l", 16, "Длина генерируемого пароля")
		cmd.Flags().StringVarP(&exclude, "exclude", "e", "", "Список символов для исключения из алфавита")

		// Флаги управления наборами символов
		cmd.Flags().BoolVar(&noLowercase, "no-lowercase", false, "Исключить строчные буквы [a-z]")
		cmd.Flags().BoolVar(&includeLowercase, "include-lowercase", true, "Включить строчные буквы [a-z]")

		cmd.Flags().BoolVar(&noUppercase, "no-uppercase", false, "Исключить прописные буквы [A-Z]")
		cmd.Flags().BoolVar(&includeUppercase, "include-uppercase", true, "Включить прописные буквы [A-Z]")

		cmd.Flags().BoolVar(&noDigits, "no-digits", false, "Исключить цифры [0-9]")
		cmd.Flags().BoolVar(&includeDigits, "include-digits", true, "Включить цифры [0-9]")

		cmd.Flags().BoolVar(&noSpecials, "no-specials", false, "Исключить специальные символы")
		cmd.Flags().BoolVar(&includeSpecials, "include-specials", true, "Включить специальные символы")
	}

	rootCmd.AddCommand(fpwCmd, fkeyCmd, keyCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Вспомогательная функция для сборки Алфавита с учетом инверсных флагов CLI
func buildAlphabetFromFlags() (detpwg.Alphabet, error) {
	// Если флаг --no-... выставлен в true, он перекрывает дефолтный include-...
	cfg := detpwg.AlphabetConfig{
		Exclude:          exclude,
		IncludeLowercase: includeLowercase && !noLowercase,
		IncludeUppercase: includeUppercase && !noUppercase,
		IncludeDigits:    includeDigits && !noDigits,
		IncludeSpecial:   includeSpecials && !noSpecials,
	}

	return cfg.BuildAlphabet()
}

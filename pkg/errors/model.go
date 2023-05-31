package errors

// CustomErr - Кастомная ошибка, которую надо показать юзеру и скрыть настоящую ошибку
// Field - поле, на котором произошла ошибка
// File - файл, в котором произошла ошибка
// Err - сама ошибка
// IsNotWrite - надо ли вывести отчет об ошибке в терминал? (по умолчанию выводит)

type CustomError struct {
	CustomErr         string
	Field             string
	File              string
	Err               error
	IsNotWriteError   bool
	IsNotWriteMessage bool
}

type NotLoggingErr error

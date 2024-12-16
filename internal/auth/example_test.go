package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/mi4r/go-url-shortener/internal/auth"
)

func ExampleGenerateUserID() {
	userID := auth.GenerateUserID()
	fmt.Println(len(userID) > 0) // Проверка, что идентификатор сгенерирован
	// Output:
	// true
}

func ExampleSignUserID() {
	userID := "example-user-id"
	signature := auth.SignUserID(userID)
	fmt.Println(len(signature) > 0) // Проверка, что подпись не пустая
	// Output:
	// true
}

func ExampleSetUserCookie() {
	userID := "example-user-id"
	w := httptest.NewRecorder()
	auth.SetUserCookie(w, userID)

	resp := w.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()

	if len(cookies) > 0 {
		fmt.Println(cookies[0].Name)  // Имя куки
		fmt.Println(cookies[0].Value) // Значение куки
	} else {
		fmt.Println("No cookies set")
	}
	// Output:
	// user_id
	// example-user-id|64a3d29d7b002efbfa93eb795869918e97b3251874aded4b7735ee22443273c8
}

func ExampleValidateUserCookie() {
	// Создаем запрос с корректной кукой
	userID := "example-user-id"
	signature := auth.SignUserID(userID)
	cookieValue := fmt.Sprintf("%s|%s", userID, signature)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: cookieValue})

	validatedUserID, isValid := auth.ValidateUserCookie(req)

	fmt.Println(isValid)         // Должно быть true
	fmt.Println(validatedUserID) // Ожидается example-user-id
	// Output:
	// true
	// example-user-id
}

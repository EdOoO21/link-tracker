package scrapper

import (
	"time"

	"github.com/gin-gonic/gin"
	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/repository"
	scrapperhttp "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/scrapper/http"
)

func main() {
	logger := logs.NewLogger()
	repo := repository.NewRepo()

	scrapperService := scrapper.NewScrapper(logger, repo)
	scrapperService.RunCron(1 * time.Minute)

	r := gin.Default()
	handler := scrapperhttp.NewHandler(logger)
	handler.RegisterRoutes(r)
	r.Run(":8080")
}

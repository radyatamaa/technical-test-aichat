package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/radyatamaa/technical-test-aichat/internal"
	"github.com/radyatamaa/technical-test-aichat/pkg/database"

	beego "github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/beego/i18n"
	"github.com/radyatamaa/technical-test-aichat/internal/domain"
	"github.com/radyatamaa/technical-test-aichat/internal/middlewares"
	"github.com/radyatamaa/technical-test-aichat/pkg/response"
	"github.com/radyatamaa/technical-test-aichat/pkg/zaplogger"

	customerVoucherRepository "github.com/radyatamaa/technical-test-aichat/internal/customer_voucher/repository"
	customerVoucherBookRepository "github.com/radyatamaa/technical-test-aichat/internal/customer_voucher_book/repository"
	purchaseTransactionRepository "github.com/radyatamaa/technical-test-aichat/internal/purchase_transaction/repository"

	customerHandler "github.com/radyatamaa/technical-test-aichat/internal/customer/delivery/http/v1"
	customerRepository "github.com/radyatamaa/technical-test-aichat/internal/customer/repository"
	customerUsecase "github.com/radyatamaa/technical-test-aichat/internal/customer/usecase"
)

// @title Api Gateway V1
// @version v1
// @contact.name radyatama
// @contact.email mohradyatama24@gmail.com
// @description api "API Gateway v1"
// @BasePath /api
// @query.collection.format multi

func main() {
	err := beego.LoadAppConfig("ini", "conf/app.ini")
	if err != nil {
		panic(err)
	}
	// global execution timeout
	serverTimeout := beego.AppConfig.DefaultInt64("serverTimeout", 60)
	// global execution timeout
	requestTimeout := beego.AppConfig.DefaultInt("executionTimeout", 5)
	// web hook to slack error log
	slackWebHookUrl := beego.AppConfig.DefaultString("slackWebhookUrlLog", "")
	// app version
	appVersion := beego.AppConfig.DefaultString("version", "1")
	// log path
	logPath := beego.AppConfig.DefaultString("logPath", "./logs/api_gateway_service.log")
	// init data
	initData := beego.AppConfig.DefaultString("initData", "true")

	// database initialization
	db := database.DB()

	// language
	lang := beego.AppConfig.DefaultString("lang", "en|id")
	languages := strings.Split(lang, "|")
	for _, value := range languages {
		if err := i18n.SetMessage(value, "./conf/"+value+".ini"); err != nil {
			panic("Failed to set message file for l10n")
		}
	}

	// global execution timeout to second
	timeoutContext := time.Duration(requestTimeout) * time.Second

	// beego config
	beego.BConfig.Log.AccessLogs = false
	beego.BConfig.Log.EnableStaticLogs = false
	beego.BConfig.Listen.ServerTimeOut = serverTimeout

	// zap logger
	zapLog := zaplogger.NewZapLogger(logPath, slackWebHookUrl)

	if beego.BConfig.RunMode == "dev" {
		// db auto migrate dev environment
		if err := db.AutoMigrate(
			&domain.Customer{},
			&domain.CustomerVoucher{},
			&domain.CustomerVoucherBook{},
			&domain.PurchaseTransaction{},
		); err != nil {
			panic(err)
		}

		// static files swagger
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	if initData == "true" {
		domain.SeederData(db)
	}
	if beego.BConfig.RunMode != "prod" {
		// static files swagger
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// middleware init
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowMethods:    []string{http.MethodGet, http.MethodPost},
		AllowAllOrigins: true,
	}))

	beego.InsertFilterChain("*", middlewares.RequestID())
	beego.InsertFilterChain("/api/*", middlewares.BodyDumpWithConfig(middlewares.NewAccessLogMiddleware(zapLog, appVersion).Logger()))

	// health check
	beego.Get("/health", func(ctx *beegoContext.Context) {
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(beego.M{"status": "alive"}, beego.BConfig.RunMode != "prod", false)
	})

	// default error handler
	beego.ErrorController(&response.ErrorController{})

	// init repository
	customerRepo := customerRepository.NewMysqlCustomerRepository(db, zapLog)
	customerVoucherRepo := customerVoucherRepository.NewMysqlCustomerVoucherRepository(db, zapLog)
	customerVoucherBookRepo := customerVoucherBookRepository.NewMysqlCCustomerVoucherBookRepository(db, zapLog)
	purchaseTransactionRepo := purchaseTransactionRepository.NewPurchaseTransactionRepository(db, zapLog)

	// init usecase
	customerUcase := customerUsecase.NewCustomerUseCase(timeoutContext,
		customerRepo,
		customerVoucherRepo,
		customerVoucherBookRepo,
		purchaseTransactionRepo,
		zapLog)

	// init handler
	customerHandler.NewCustomerHandler(customerUcase, zapLog)

	// default error handler
	beego.ErrorController(&internal.BaseController{})

	beego.Run()
}

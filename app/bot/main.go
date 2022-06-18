package main

import (
	"database/sql"
	"fmt"
	"log"

	migrate "github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // export migrations from files
	"moul.io/zapgorm2"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kelseyhightower/envconfig"
	"github.com/kranikitao/fxhash-telegram-bot/src/artcollector"
	"github.com/kranikitao/fxhash-telegram-bot/src/chat"
	"github.com/kranikitao/fxhash-telegram-bot/src/fxhash"
	"github.com/kranikitao/fxhash-telegram-bot/src/sender"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	TGToken    string `envconfig:"TG_TOKEN"`
	DBName     string `envconfig:"DB_NAME"`
	DBHost     string `envconfig:"DB_HOST"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBUser     string `envconfig:"DB_USER"`
	DBPort     uint   `envconfig:"DB_PORT"`
}

func main() {
	config := loadConfigurationFromEnvironments("FXBOT")
	connectionLogger := newLogger("conncection")
	dbConnection := connectToDatabase(config, connectionLogger)
	defer dbConnection.Close()

	gormLogger := zapgorm2.New(connectionLogger)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbConnection,
	}), &gorm.Config{Logger: gormLogger})
	if err != nil {
		log.Panic(err.Error())
	}

	botLogger := newLogger("bot")

	bot, err := tgbotapi.NewBotAPI(config.TGToken)
	if err != nil {
		botLogger.Panic("can't connect to bot api", zap.Error(err))
	}

	setCommandsRequest := tgbotapi.NewSetMyCommandsWithScope(
		tgbotapi.NewBotCommandScopeDefault(),
		tgbotapi.BotCommand{Command: chat.CommandSubscribeArtist, Description: "Subscribe to artist"},
		tgbotapi.BotCommand{Command: chat.CommandSubscribeFree, Description: "Subscribe to zero cost generatives"},
		tgbotapi.BotCommand{Command: chat.CommandUnsubscribe, Description: "Unsubscribe"},
		tgbotapi.BotCommand{Command: chat.CommandCancel, Description: "Cancel operation"},
	)

	_, err = bot.Request(setCommandsRequest)

	if err != nil {
		log.Panic(err)
	}
	go artcollector.New(newLogger("collector"), fxhash.New(), gormDB).Collect()
	go sender.New(newLogger("sender"), bot, gormDB).Start()

	chat.New(bot, botLogger, gormDB).Start()
}

func connectToDatabase(config *Config, logger *zap.Logger) *sql.DB {
	connectionString := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable port=%d", config.DBHost, config.DBUser, config.DBName, config.DBPassword, config.DBPort)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Panic(err.Error())
	}

	migrateIt(db, config, logger)

	return db
}

func migrateIt(db *sql.DB, config *Config, logger *zap.Logger) {
	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://src/migrations", config.DBName, driver)

	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		logger.Warn(
			"sql migration result",
			zap.Error(err),
		)
	} else {
		logger.Info("Schema was migrated")
	}
}

func newLogger(name string) *zap.Logger {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	logger, err := config.Build()

	if err != nil {
		log.Fatal(err)
	}
	logger = logger.WithOptions().Named(name)
	defer logger.Sync()

	return logger
}

func loadConfigurationFromEnvironments(prefix string) *Config {
	config := &Config{}
	err := envconfig.Process(prefix, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return config
}

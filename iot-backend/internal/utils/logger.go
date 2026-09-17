// internal/utils/logger.go
// VERSION CORRIGÉE v3.0 — Logger structuré via log/slog (Go 1.21+)
//
// CORRECTIONS v2 → v3 :
//
//   [FIX-CRITIQUE-1]  Logger.Info/Error/Warn utilisaient log.Println(msg, keys)
//                     Les clés key-value étaient affichées comme un slice Go :
//                     "INFO: Starting server [[error foo]]" — impossible à parser.
//                     → Remplacement par log/slog (stdlib Go 1.21) avec TextHandler
//                       en développement et JSONHandler en production.
//                       Sortie structurée : time=... level=INFO msg=... key=val
//
//   [FIX-CRITIQUE-2]  RWMutex utilisé alors que log.Logger.Println n'est pas
//                     protégé en lecture concurrente (il écrit dans io.Writer).
//                     → Supprimé : slog.Logger est thread-safe nativement.
//
//   [FIX-MOYEN-1]    InitLogger() et GetLogger() pouvaient être appelés dans
//                    n'importe quel ordre ; double initialisation silencieuse.
//                    → sync.Once garantit une seule initialisation.
//
//   [FIX-MOYEN-2]    Niveau de log non configurable (DEBUG toujours actif).
//                    → Niveau configurable via LOG_LEVEL (debug/info/warn/error).
//
//   [FIX-MINEUR-1]   Format non configurable (texte ou JSON selon ENV).
//                    → LOG_FORMAT=json pour production (Loki, ELK, CloudWatch).

package utils

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Logger encapsule slog.Logger avec une interface compatible
// avec l'API existante du projet (Info, Warn, Error avec key-value pairs).
type Logger struct {
	sl *slog.Logger
}

var (
	instance     *Logger
	instanceOnce sync.Once
)

// InitLogger initialise le logger global.
// Idempotent via sync.Once — les appels successifs sont sans effet.
//
// Variables d'environnement :
//   - LOG_LEVEL  : debug | info | warn | error (défaut: info)
//   - LOG_FORMAT : text | json               (défaut: text, json si ENV=production)
func InitLogger() {
	instanceOnce.Do(func() {
		level := parseLevel(os.Getenv("LOG_LEVEL"))

		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: level == slog.LevelDebug,
		}

		var handler slog.Handler
		format := strings.ToLower(os.Getenv("LOG_FORMAT"))
		if format == "json" || os.Getenv("ENV") == "production" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		instance = &Logger{sl: slog.New(handler)}
	})
}

// GetLogger retourne le logger global, en l'initialisant si nécessaire.
func GetLogger() *Logger {
	instanceOnce.Do(InitLogger)
	return instance
}

// Info logue un message de niveau INFO avec des paires clé-valeur optionnelles.
//
// Utilisation :
//
//	logger.Info("Serveur démarré", "port", 8443, "tls", true)
//	→ time=... level=INFO msg="Serveur démarré" port=8443 tls=true
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.sl.Info(msg, keysAndValues...)
}

// Warn logue un message de niveau WARN avec des paires clé-valeur optionnelles.
func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.sl.Warn(msg, keysAndValues...)
}

// Error logue un message de niveau ERROR avec des paires clé-valeur optionnelles.
func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.sl.Error(msg, keysAndValues...)
}

// Debug logue un message de niveau DEBUG avec des paires clé-valeur optionnelles.
// Visible uniquement si LOG_LEVEL=debug.
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.sl.Debug(msg, keysAndValues...)
}

// With retourne un nouveau Logger avec des attributs permanents attachés.
// Utile pour contextualiser les logs d'un sous-système (ex: device_id).
//
// Exemple :
//
//	deviceLogger := logger.With("device_id", d.ID, "address", d.Address)
//	deviceLogger.Info("Enregistrement") // → ... device_id=IOT_0001 address=0x...
func (l *Logger) With(keysAndValues ...any) *Logger {
	return &Logger{sl: l.sl.With(keysAndValues...)}
}

// parseLevel convertit une string en slog.Level.
// Valeurs acceptées (insensible à la casse) : debug, info, warn, error.
// Valeur par défaut : slog.LevelInfo.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

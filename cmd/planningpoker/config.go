package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const envListenAddr = "PLANNINGPOKER_LISTEN_ADDR"

// defaultListenAddr binds all interfaces; an unset or empty setting uses this default.
const defaultListenAddr = ":8080"

const envShutdownTimeout = "PLANNINGPOKER_SHUTDOWN_TIMEOUT"

const envRoomGracePeriod = "PLANNINGPOKER_ROOM_GRACE_PERIOD"

// defaultRoomGracePeriod retains empty rooms through brief disconnections.
const defaultRoomGracePeriod = 5 * time.Minute

const envMaxRooms = "PLANNINGPOKER_MAX_ROOMS"

const envMaxConnectionsPerRoom = "PLANNINGPOKER_MAX_CONNECTIONS_PER_ROOM"

const envMaxParticipantsPerRoom = "PLANNINGPOKER_MAX_PARTICIPANTS_PER_ROOM"

const envMessageRate = "PLANNINGPOKER_MESSAGE_RATE"

const envLegalDir = "PLANNINGPOKER_LEGAL_DIR"

// Default capacity ceilings for a small deployment.
const (
	defaultMaxRooms               = 50
	defaultMaxConnectionsPerRoom  = 40
	defaultMaxParticipantsPerRoom = 20
	defaultMessageRate            = 10
)

// messageBurstFactor allows a short burst at twice the sustained rate.
const messageBurstFactor = 2

// defaultShutdownTimeout leaves time before Compose sends SIGKILL at ten seconds.
const defaultShutdownTimeout = 5 * time.Second

// config is read once at startup and passed to the application components.
type config struct {
	// ListenAddr is a host:port address. An empty host binds all interfaces.
	// Port 0 requests an available port; invalid addresses fail at startup.
	ListenAddr string

	// ShutdownTimeout bounds the wait for HTTP and WebSocket shutdown. It must be
	// positive and should stay below the container stop grace period. An expired
	// HTTP shutdown currently returns an error, causing a nonzero exit.
	ShutdownTimeout time.Duration

	// RoomGracePeriod retains a room after its last connection closes, or after
	// creation if nobody connects. It must be positive. Connected rooms do not expire;
	// longer values retain abandoned room data longer.
	RoomGracePeriod time.Duration

	// MaxRooms limits rooms held in memory and must be positive. At capacity, new
	// rooms are refused while existing rooms remain accessible.
	MaxRooms int

	// MaxConnectionsPerRoom limits attached sockets, including multiple tabs per
	// participant. It must be positive; extra connections are refused.
	MaxConnectionsPerRoom int

	// MaxParticipantsPerRoom limits seats, including away participants. It must be
	// positive. Returning to an existing seat does not consume another seat.
	MaxParticipantsPerRoom int

	// MessageRate is the sustained message allowance per second per connection.
	// It must be positive; a burst of twice this value is allowed. Excess messages
	// are refused and sustained flooding can close the connection. This is not an
	// inactivity timeout.
	MessageRate int

	// LegalDir is the directory holding the operator's privacy.html and imprint.html.
	// Empty, the default, leaves legal notices switched off and reads no files at all.
	// A relative path is resolved against the process working directory. When it is
	// set, both documents must be present at startup or the process refuses to start;
	// they are read once, so edited text takes effect on the next restart.
	LegalDir string
}

// MessageBurst returns the per-connection burst allowance.
func (c config) MessageBurst() int { return c.MessageRate * messageBurstFactor }

// loadConfig reads and validates startup settings. getenv is injected for tests.
func loadConfig(getenv func(string) string) (config, error) {
	cfg := config{
		ListenAddr:             defaultListenAddr,
		ShutdownTimeout:        defaultShutdownTimeout,
		RoomGracePeriod:        defaultRoomGracePeriod,
		MaxRooms:               defaultMaxRooms,
		MaxConnectionsPerRoom:  defaultMaxConnectionsPerRoom,
		MaxParticipantsPerRoom: defaultMaxParticipantsPerRoom,
		MessageRate:            defaultMessageRate,
	}

	if raw := getenv(envListenAddr); raw != "" {
		if err := validateListenAddr(raw); err != nil {
			return config{}, fmt.Errorf("%s=%q is not a valid listen address: %w", envListenAddr, raw, err)
		}
		cfg.ListenAddr = raw
	}

	if raw := getenv(envShutdownTimeout); raw != "" {
		timeout, err := parsePositiveDuration(raw)
		if err != nil {
			return config{}, fmt.Errorf("%s=%q is not a valid shutdown timeout: %w", envShutdownTimeout, raw, err)
		}
		cfg.ShutdownTimeout = timeout
	}

	if raw := getenv(envRoomGracePeriod); raw != "" {
		grace, err := parsePositiveDuration(raw)
		if err != nil {
			return config{}, fmt.Errorf("%s=%q is not a valid room grace period: %w", envRoomGracePeriod, raw, err)
		}
		cfg.RoomGracePeriod = grace
	}

	for _, limit := range []struct {
		variable string
		what     string
		into     *int
	}{
		{envMaxRooms, "room limit", &cfg.MaxRooms},
		{envMaxConnectionsPerRoom, "connection limit", &cfg.MaxConnectionsPerRoom},
		{envMaxParticipantsPerRoom, "participant limit", &cfg.MaxParticipantsPerRoom},
		{envMessageRate, "message rate", &cfg.MessageRate},
	} {
		raw := getenv(limit.variable)
		if raw == "" {
			continue
		}
		value, err := parsePositiveCount(raw)
		if err != nil {
			return config{}, fmt.Errorf("%s=%q is not a valid %s: %w", limit.variable, raw, limit.what, err)
		}
		*limit.into = value
	}

	// The path is not inspected here. Whether the directory holds usable documents is
	// decided when they are loaded, so one place reports every reason they were refused.
	cfg.LegalDir = getenv(envLegalDir)

	return cfg, nil
}

// parsePositiveCount rejects zero, negative and non-integer limits.
func parsePositiveCount(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("expected a whole number, for example %q", "50")
	}
	if value <= 0 {
		return 0, fmt.Errorf("must be greater than zero, got %d", value)
	}
	return value, nil
}

// parsePositiveDuration rejects zero, negative and malformed durations.
func parsePositiveDuration(raw string) (time.Duration, error) {
	timeout, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf(`expected a duration such as "5s", "1500ms" or "1m30s"`)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("must be greater than zero, got %s", timeout)
	}
	return timeout, nil
}

// validateListenAddr checks address syntax before binding, so errors can name the
// setting.
func validateListenAddr(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("expected the form host:port, for example %q", defaultListenAddr)
	}

	number, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port %q is not a number", port)
	}
	if number < 0 || number > 65535 {
		return fmt.Errorf("port %d is outside the valid range 0-65535", number)
	}

	// Host validation and name resolution are left to net.Listen.
	return nil
}

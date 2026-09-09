package main

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// env builds a getenv function over a fixed map, so no test touches the real
// environment of the test process.
func env(pairs map[string]string) func(string) string {
	return func(key string) string { return pairs[key] }
}

func TestUnsetVariableYieldsTheDefault(t *testing.T) {
	for _, pairs := range []map[string]string{
		{},                          // variable absent
		{envListenAddr: ""},         // variable present but empty
		{"SOMETHING_ELSE": ":9999"}, // an unrelated variable
	} {
		cfg, err := loadConfig(env(pairs))
		if err != nil {
			t.Fatalf("loadConfig(%v) returned an error: %v", pairs, err)
		}
		if cfg.ListenAddr != defaultListenAddr {
			t.Errorf("loadConfig(%v).ListenAddr = %q, want the default %q", pairs, cfg.ListenAddr, defaultListenAddr)
		}
	}
}

func TestVariableOverridesTheDefault(t *testing.T) {
	for _, addr := range []string{":9000", "127.0.0.1:8080", "0.0.0.0:80", "localhost:3000"} {
		cfg, err := loadConfig(env(map[string]string{envListenAddr: addr}))
		if err != nil {
			t.Fatalf("loadConfig with %q returned an error: %v", addr, err)
		}
		if cfg.ListenAddr != addr {
			t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, addr)
		}
	}
}

func TestInvalidValueIsAnErrorNamingVariableAndValue(t *testing.T) {
	for _, tc := range []struct{ variable, value string }{
		{envListenAddr, "8080"},
		{envListenAddr, "no-colon-here"},
		{envListenAddr, ":not-a-number"},
		{envListenAddr, ":70000"},
		{envListenAddr, ":-1"},
		{envShutdownTimeout, "soon"},
		{envShutdownTimeout, "5"},      // a bare number has no unit
		{envShutdownTimeout, "0"},      // zero is ambiguous, so it is refused
		{envShutdownTimeout, "0s"},     // likewise
		{envShutdownTimeout, "-5s"},    // a negative budget means nothing
		{envShutdownTimeout, "5 secs"}, // not a duration Go can read
		{envRoomGracePeriod, "soon"},
		{envRoomGracePeriod, "5"},   // a bare number has no unit
		{envRoomGracePeriod, "0"},   // there is deliberately no "never expire"
		{envRoomGracePeriod, "0m"},  // likewise
		{envRoomGracePeriod, "-5m"}, // a negative lifetime means nothing
	} {
		_, err := loadConfig(env(map[string]string{tc.variable: tc.value}))
		if err == nil {
			t.Errorf("%s=%q returned no error, want one — an invalid value must never fall back to the default",
				tc.variable, tc.value)
			continue
		}
		if !strings.Contains(err.Error(), tc.variable) {
			t.Errorf("error for %s=%q does not name the variable: %v", tc.variable, tc.value, err)
		}
		if !strings.Contains(err.Error(), tc.value) {
			t.Errorf("error for %s=%q does not quote the offending value: %v", tc.variable, tc.value, err)
		}
	}
}

func TestUnsetShutdownTimeoutYieldsTheDefault(t *testing.T) {
	for _, pairs := range []map[string]string{
		{},
		{envShutdownTimeout: ""},
		{envListenAddr: ":9000"}, // the other variable set, this one not
	} {
		cfg, err := loadConfig(env(pairs))
		if err != nil {
			t.Fatalf("loadConfig(%v) returned an error: %v", pairs, err)
		}
		if cfg.ShutdownTimeout != defaultShutdownTimeout {
			t.Errorf("loadConfig(%v).ShutdownTimeout = %s, want the default %s",
				pairs, cfg.ShutdownTimeout, defaultShutdownTimeout)
		}
	}
}

func TestShutdownTimeoutVariableOverridesTheDefault(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  time.Duration
	}{
		{"1ms", time.Millisecond},
		{"3s", 3 * time.Second},
		{"1500ms", 1500 * time.Millisecond},
		{"1m30s", 90 * time.Second},
	} {
		cfg, err := loadConfig(env(map[string]string{envShutdownTimeout: tc.value}))
		if err != nil {
			t.Fatalf("loadConfig with %q returned an error: %v", tc.value, err)
		}
		if cfg.ShutdownTimeout != tc.want {
			t.Errorf("ShutdownTimeout for %q = %s, want %s", tc.value, cfg.ShutdownTimeout, tc.want)
		}
	}
}

func TestDefaultShutdownTimeoutStaysBelowTheContainerGracePeriod(t *testing.T) {
	// Both `docker stop` and Compose wait ten seconds after SIGTERM before sending
	// SIGKILL. A default at or above that would let the process be killed in the
	// middle of the cleanup this timeout exists to make possible.
	const containerGracePeriod = 10 * time.Second
	if defaultShutdownTimeout >= containerGracePeriod {
		t.Errorf("defaultShutdownTimeout is %s, which is not below the container grace period of %s",
			defaultShutdownTimeout, containerGracePeriod)
	}
}

func TestUnsetRoomGracePeriodYieldsTheDefault(t *testing.T) {
	for _, pairs := range []map[string]string{
		{},
		{envRoomGracePeriod: ""},
		{envListenAddr: ":9000", envShutdownTimeout: "3s"}, // the others set, this one not
	} {
		cfg, err := loadConfig(env(pairs))
		if err != nil {
			t.Fatalf("loadConfig(%v) returned an error: %v", pairs, err)
		}
		if cfg.RoomGracePeriod != defaultRoomGracePeriod {
			t.Errorf("loadConfig(%v).RoomGracePeriod = %s, want the default %s",
				pairs, cfg.RoomGracePeriod, defaultRoomGracePeriod)
		}
	}
}

func TestRoomGracePeriodVariableOverridesTheDefault(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  time.Duration
	}{
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"1h30m", 90 * time.Minute},
	} {
		cfg, err := loadConfig(env(map[string]string{envRoomGracePeriod: tc.value}))
		if err != nil {
			t.Fatalf("loadConfig with %q returned an error: %v", tc.value, err)
		}
		if cfg.RoomGracePeriod != tc.want {
			t.Errorf("RoomGracePeriod for %q = %s, want %s", tc.value, cfg.RoomGracePeriod, tc.want)
		}
	}
}

// capacityLimits pairs each capacity variable with the field it fills and its
// default, so that the tests below cover all four without repeating themselves once
// per variable.
var capacityLimits = []struct {
	variable string
	value    func(config) int
	def      int
}{
	{envMaxRooms, func(c config) int { return c.MaxRooms }, defaultMaxRooms},
	{envMaxConnectionsPerRoom, func(c config) int { return c.MaxConnectionsPerRoom }, defaultMaxConnectionsPerRoom},
	{envMaxParticipantsPerRoom, func(c config) int { return c.MaxParticipantsPerRoom }, defaultMaxParticipantsPerRoom},
	{envMessageRate, func(c config) int { return c.MessageRate }, defaultMessageRate},
}

func TestUnsetCapacityLimitYieldsTheDefault(t *testing.T) {
	for _, limit := range capacityLimits {
		for _, pairs := range []map[string]string{
			{},
			{limit.variable: ""},
			{envListenAddr: ":9000"}, // an unrelated variable set, this one not
		} {
			cfg, err := loadConfig(env(pairs))
			if err != nil {
				t.Fatalf("loadConfig(%v) returned an error: %v", pairs, err)
			}
			if got := limit.value(cfg); got != limit.def {
				t.Errorf("%s with %v = %d, want the default %d", limit.variable, pairs, got, limit.def)
			}
		}
	}
}

func TestCapacityLimitVariableOverridesTheDefault(t *testing.T) {
	for _, limit := range capacityLimits {
		for _, want := range []int{1, 7, 1000} {
			cfg, err := loadConfig(env(map[string]string{limit.variable: strconv.Itoa(want)}))
			if err != nil {
				t.Fatalf("loadConfig with %s=%d returned an error: %v", limit.variable, want, err)
			}
			if got := limit.value(cfg); got != want {
				t.Errorf("%s=%d gave %d", limit.variable, want, got)
			}
		}
	}
}

func TestInvalidCapacityLimitIsAnErrorNamingVariableAndValue(t *testing.T) {
	// Every one of these is refused for the same reason, so each boundary is covered
	// once across the four variables rather than once per variable: what is being
	// tested is the shared parsing, and every variable goes through it.
	for i, value := range []string{
		"0",    // there is deliberately no value meaning "no limit"
		"-1",   // a negative ceiling means nothing
		"many", // not a number at all
		"1.5",  // half a room is not a room
		"10 ",  // a stray space is a typo, not a value to guess at
		"1e3",  // scientific notation is not a whole number here
		"50rooms",
	} {
		limit := capacityLimits[i%len(capacityLimits)]
		_, err := loadConfig(env(map[string]string{limit.variable: value}))
		if err == nil {
			t.Errorf("%s=%q returned no error, want one — an invalid value must never fall back to the default",
				limit.variable, value)
			continue
		}
		if !strings.Contains(err.Error(), limit.variable) {
			t.Errorf("error for %s=%q does not name the variable: %v", limit.variable, value, err)
		}
		if !strings.Contains(err.Error(), value) {
			t.Errorf("error for %s=%q does not quote the offending value: %v", limit.variable, value, err)
		}
	}
}

func TestUnsetLegalDirectoryLeavesNoticesOff(t *testing.T) {
	// Legal notices are opt-in. A configuration that says nothing about them must
	// leave them switched off, because the alternative would be a default
	// installation that fails to start unless it is given documents.
	for _, pairs := range []map[string]string{
		{},
		{envLegalDir: ""},
		{envListenAddr: ":9000", envRoomGracePeriod: "1m"}, // others set, this one not
	} {
		cfg, err := loadConfig(env(pairs))
		if err != nil {
			t.Fatalf("loadConfig(%v) returned an error: %v", pairs, err)
		}
		if cfg.LegalDir != "" {
			t.Errorf("loadConfig(%v).LegalDir = %q, want it empty", pairs, cfg.LegalDir)
		}
	}
}

func TestLegalDirectoryIsTakenVerbatim(t *testing.T) {
	// Whether the path is usable is decided when the documents are loaded, so the
	// configuration passes on exactly what the operator wrote — including a
	// relative path, which is resolved against the process working directory.
	for _, dir := range []string{"/legal", "./legal", "legal", "/srv/planningpoker/notices"} {
		cfg, err := loadConfig(env(map[string]string{envLegalDir: dir}))
		if err != nil {
			t.Fatalf("loadConfig with %q returned an error: %v", dir, err)
		}
		if cfg.LegalDir != dir {
			t.Errorf("LegalDir = %q, want %q", cfg.LegalDir, dir)
		}
	}
}

func TestMessageBurstIsAboveTheSustainedRate(t *testing.T) {
	// The burst exists so that a handful of intents arriving together — a page
	// reconnecting, or a vote followed at once by a reveal — is not mistaken for a
	// flood. A burst equal to the rate would defeat that.
	cfg, err := loadConfig(env(map[string]string{}))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.MessageBurst() <= cfg.MessageRate {
		t.Errorf("MessageBurst() = %d, which is not above the sustained rate of %d",
			cfg.MessageBurst(), cfg.MessageRate)
	}
}

func TestTheSettingsAreIndependent(t *testing.T) {
	// Setting one must not disturb the others: a configuration where changing the
	// grace period silently reset the listen address would be very hard to diagnose.
	cfg, err := loadConfig(env(map[string]string{
		envListenAddr:             "127.0.0.1:9999",
		envShutdownTimeout:        "2s",
		envRoomGracePeriod:        "90s",
		envMaxRooms:               "11",
		envMaxConnectionsPerRoom:  "12",
		envMaxParticipantsPerRoom: "13",
		envMessageRate:            "14",
		envLegalDir:               "/legal",
	}))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.ListenAddr != "127.0.0.1:9999" {
		t.Errorf("ListenAddr = %q", cfg.ListenAddr)
	}
	if cfg.ShutdownTimeout != 2*time.Second {
		t.Errorf("ShutdownTimeout = %s", cfg.ShutdownTimeout)
	}
	if cfg.RoomGracePeriod != 90*time.Second {
		t.Errorf("RoomGracePeriod = %s", cfg.RoomGracePeriod)
	}
	if cfg.MaxRooms != 11 {
		t.Errorf("MaxRooms = %d", cfg.MaxRooms)
	}
	if cfg.MaxConnectionsPerRoom != 12 {
		t.Errorf("MaxConnectionsPerRoom = %d", cfg.MaxConnectionsPerRoom)
	}
	if cfg.MaxParticipantsPerRoom != 13 {
		t.Errorf("MaxParticipantsPerRoom = %d", cfg.MaxParticipantsPerRoom)
	}
	if cfg.MessageRate != 14 {
		t.Errorf("MessageRate = %d", cfg.MessageRate)
	}
	if cfg.LegalDir != "/legal" {
		t.Errorf("LegalDir = %q", cfg.LegalDir)
	}
}

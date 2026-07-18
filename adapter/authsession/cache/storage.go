package cache

import (
	"fmt"
	"time"

	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
	"github.com/webcore-go/webcore/port/auth"
)

type MemoryCacheSessionStore struct {
	Config config.AuthConfig
	Memory port.ICacheMemory
}

func MemoryCacheSessionBackend(config config.AuthConfig, context *core.AppContext) (*MemoryCacheSessionStore, error) {
	libMem, ok := context.GetSingletonInstance(config.Session.Backend)
	if !ok {
		return nil, fmt.Errorf("Memory Session cannot be loaded, backend %s not found", config.Session.Backend)
	}

	memory := libMem.(port.ICacheMemory)
	return &MemoryCacheSessionStore{
		Config: config,
		Memory: memory,
	}, nil
}

func (m *MemoryCacheSessionStore) save(loginInfo *auth.UserLoginInfo, expires time.Duration, refresh time.Duration) error {
	if loginInfo == nil {
		return nil
	}

	// Simpan dengan ke Access Token
	memkey1 := "authsess_at_" + *loginInfo.AccessToken
	err := m.Memory.Set(memkey1, *loginInfo, expires)
	if err != nil {
		return err
	}

	// Simpan backup dengan ke Refresh Token
	memkey2 := "authsess_rt_" + *loginInfo.RefreshToken
	err = m.Memory.Set(memkey2, *loginInfo, refresh)
	if err != nil {
		return err
	}

	// Simpan backup dengan ke Username
	memkey3 := "authsess_usr_" + loginInfo.Username
	err = m.Memory.Set(memkey3, *loginInfo, refresh)
	if err != nil {
		return err
	}

	return nil
}

func (m *MemoryCacheSessionStore) Save(loginInfo *auth.UserLoginInfo) error {
	return m.save(loginInfo, loginInfo.ExpiresIn+60*time.Second, loginInfo.RefreshIn+60*time.Second)
}

func (m *MemoryCacheSessionStore) Delete(loginInfo *auth.UserLoginInfo) error {
	return m.save(loginInfo, 1*time.Second, 1*time.Second)
}

// Gunakan expiration (refreshIn) yang lama untuk Refresh Token,
// jika generate baru maka refresh token tidak pernah invalidate dan dengan begitu user akan selalu bisa login selamanya (jika tidak logout manual)
func (m *MemoryCacheSessionStore) Refresh(oldAccessToken string, oldRefreshToken string, loginInfo *auth.UserLoginInfo, refreshIn time.Duration) error {
	if loginInfo == nil {
		return nil
	}

	// Segera buat expired Access Token lama
	memkey0 := "authsess_at_" + oldAccessToken
	m.Memory.Set(memkey0, *loginInfo, 1*time.Second)

	// Segera buat expired Refresh Token lama
	memkey1 := "authsess_rt_" + oldRefreshToken
	m.Memory.Set(memkey1, *loginInfo, 1*time.Second)

	accessTokenExp := loginInfo.ExpiresIn + 60*time.Second

	return m.save(loginInfo, accessTokenExp, refreshIn)
}

func (m *MemoryCacheSessionStore) GetByAccessToken(accessToken string) (*auth.UserLoginInfo, error) {
	var value auth.UserLoginInfo
	memkey1 := "authsess_at_" + accessToken
	if ok := m.Memory.Get(memkey1, &value); !ok {
		return nil, fmt.Errorf("Gagal Mengambil Login Info dari Cache Memory")
	}

	return &value, nil
}

func (m *MemoryCacheSessionStore) GetByRefreshToken(refreshToken string) (*auth.UserLoginInfo, error) {
	var value auth.UserLoginInfo
	memkey2 := "authsess_rt_" + refreshToken
	if ok := m.Memory.Get(memkey2, &value); !ok {
		return nil, fmt.Errorf("Gagal Mengambil Login Info dari Cache Memory")
	}

	return &value, nil
}

func (m *MemoryCacheSessionStore) GetByUsername(username string) (*auth.UserLoginInfo, error) {
	var value auth.UserLoginInfo
	memkey2 := "authsess_usr_" + username
	logger.Debug("Cek User sudah punya session", "user", username)
	if ok := m.Memory.Get(memkey2, &value); !ok {
		return nil, fmt.Errorf("Gagal Mengambil Login Info dari Cache Memory")
	}

	return &value, nil
}

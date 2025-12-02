package offlinedata

import (
	"context"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
)

// Entry 离线数据条目
type Entry struct {
	RoleID    uint64
	DataType  protocol.OfflineDataType
	Payload   []byte
	Version   uint32
	UpdatedAt int64
	Dirty     bool
}

// Manager 负责 PublicActor 侧的离线数据缓存与持久化
type Manager struct {
	data map[uint64]map[protocol.OfflineDataType]*Entry
}

// NewManager 创建离线数据管理器
func NewManager() *Manager {
	return &Manager{
		data: make(map[uint64]map[protocol.OfflineDataType]*Entry),
	}
}

// LoadAll 从数据库加载所有离线数据
func (m *Manager) LoadAll(_ context.Context) error {
	records, err := database.GetAllOfflineData()
	if err != nil {
		return err
	}
	for _, record := range records {
		m.setEntry(record.RoleID, protocol.OfflineDataType(record.DataType), record.Data, record.Version, record.UpdatedAt, false)
	}
	log.Infof("OfflineDataManager loaded %d records", len(records))
	return nil
}

// UpdateProto 写入 proto.Message 数据
func (m *Manager) UpdateProto(roleID uint64, dataType protocol.OfflineDataType, message proto.Message, updatedAt int64, version uint32) error {
	if message == nil {
		return nil
	}
	payload, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return m.UpdateRaw(roleID, dataType, payload, updatedAt, version)
}

// UpdateRaw 写入原始数据
func (m *Manager) UpdateRaw(roleID uint64, dataType protocol.OfflineDataType, payload []byte, updatedAt int64, version uint32) error {
	if payload == nil {
		return nil
	}
	if updatedAt <= 0 {
		updatedAt = servertime.UnixMilli()
	}
	if version == 0 {
		version = 1
	}
	entry := m.getOrCreateEntry(roleID, dataType)
	entry.Payload = payload
	entry.Version = version
	entry.UpdatedAt = updatedAt
	entry.Dirty = true
	return nil
}

// GetProto 将离线数据反序列化到 target
func (m *Manager) GetProto(roleID uint64, dataType protocol.OfflineDataType, target proto.Message) (bool, error) {
	entry, ok := m.getEntry(roleID, dataType)
	if !ok || entry.Payload == nil {
		return false, nil
	}
	if err := proto.Unmarshal(entry.Payload, target); err != nil {
		return false, err
	}
	return true, nil
}

// GetRaw 返回原始字节（不会复制，调用方谨慎修改）
func (m *Manager) GetRaw(roleID uint64, dataType protocol.OfflineDataType) ([]byte, bool) {
	entry, ok := m.getEntry(roleID, dataType)
	if !ok {
		return nil, false
	}
	return entry.Payload, true
}

// FlushDirty 将所有 dirty 条目落库
func (m *Manager) FlushDirty(_ context.Context) {
	dirtyCount := 0
	for roleID, typeMap := range m.data {
		for dataType, entry := range typeMap {
			if !entry.Dirty {
				continue
			}
			record := &database.OfflineData{
				RoleID:    roleID,
				DataType:  uint32(dataType),
				Data:      entry.Payload,
				Version:   entry.Version,
				UpdatedAt: entry.UpdatedAt,
			}
			if err := database.UpsertOfflineData(record); err != nil {
				log.Errorf("failed to flush offline data role=%d type=%d: %v", roleID, dataType, err)
				continue
			}
			entry.Dirty = false
			dirtyCount++
		}
	}
	if dirtyCount > 0 {
		log.Debugf("OfflineDataManager flushed %d entries", dirtyCount)
	}
}

// Count 返回缓存条目数量
func (m *Manager) Count() int {
	total := 0
	for _, typeMap := range m.data {
		total += len(typeMap)
	}
	return total
}

// getEntry 获取条目
func (m *Manager) getEntry(roleID uint64, dataType protocol.OfflineDataType) (*Entry, bool) {
	typeMap, ok := m.data[roleID]
	if !ok {
		return nil, false
	}
	entry, ok := typeMap[dataType]
	return entry, ok
}

func (m *Manager) getOrCreateEntry(roleID uint64, dataType protocol.OfflineDataType) *Entry {
	typeMap, ok := m.data[roleID]
	if !ok {
		typeMap = make(map[protocol.OfflineDataType]*Entry)
		m.data[roleID] = typeMap
	}
	entry, ok := typeMap[dataType]
	if !ok {
		entry = &Entry{
			RoleID:   roleID,
			DataType: dataType,
		}
		typeMap[dataType] = entry
	}
	return entry
}

func (m *Manager) setEntry(roleID uint64, dataType protocol.OfflineDataType, payload []byte, version uint32, updatedAt int64, dirty bool) {
	entry := m.getOrCreateEntry(roleID, dataType)
	entry.Payload = payload
	entry.Version = version
	entry.UpdatedAt = updatedAt
	entry.Dirty = dirty
}

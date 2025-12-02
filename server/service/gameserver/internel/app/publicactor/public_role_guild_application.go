package publicactor

// 公会申请相关逻辑

// GuildApplication 公会申请
type GuildApplication struct {
	GuildId       uint64
	ApplicantId   uint64
	ApplicantName string
	ApplyTime     int64
}

// AddGuildApplication 添加公会申请
func (pr *PublicRole) AddGuildApplication(guildId uint64, application *GuildApplication) {
	value, _ := pr.guildApplicationMap.LoadOrStore(guildId, make([]*GuildApplication, 0))
	applications := value.([]*GuildApplication)
	applications = append(applications, application)
	pr.guildApplicationMap.Store(guildId, applications)
}

// GetGuildApplications 获取公会申请列表
func (pr *PublicRole) GetGuildApplications(guildId uint64) []*GuildApplication {
	value, ok := pr.guildApplicationMap.Load(guildId)
	if !ok {
		return nil
	}
	return value.([]*GuildApplication)
}

// RemoveGuildApplication 移除公会申请
func (pr *PublicRole) RemoveGuildApplication(guildId uint64, applicantId uint64) {
	value, ok := pr.guildApplicationMap.Load(guildId)
	if !ok {
		return
	}
	applications := value.([]*GuildApplication)
	newApplications := make([]*GuildApplication, 0)
	for _, app := range applications {
		if app.ApplicantId != applicantId {
			newApplications = append(newApplications, app)
		}
	}
	if len(newApplications) == 0 {
		pr.guildApplicationMap.Delete(guildId)
	} else {
		pr.guildApplicationMap.Store(guildId, newApplications)
	}
}

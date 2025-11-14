package entity

import (
	"context"
	"fmt"
	"strings"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
)

func (pr *PlayerRole) CheckConsume(ctx context.Context, items []*jsonconf.ItemAmount) error {
	normalized := normalizeAmounts(items)
	if len(normalized) == 0 {
		return nil
	}
	bagSys := entitysystem.GetBagSys(ctx)
	moneySys := entitysystem.GetMoneySys(ctx)

	for _, item := range normalized {
		switch item.ItemType {
		case uint32(protocol.ItemType_ItemTypeMoney):
			if moneySys == nil || moneySys.GetAmount(item.ItemId) < item.Count {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money not enough")
			}
		default:
			if bagSys == nil || !bagSys.HasItem(item.ItemId, uint32(item.Count)) {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
			}
		}
	}
	return nil
}

func (pr *PlayerRole) ApplyConsume(ctx context.Context, items []*jsonconf.ItemAmount) error {
	if err := pr.CheckConsume(ctx, items); err != nil {
		return err
	}
	normalized := normalizeAmounts(items)
	if len(normalized) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = pr.WithContext(nil)
	}
	bagSys := entitysystem.GetBagSys(ctx)
	moneySys := entitysystem.GetMoneySys(ctx)

	// 记录需要回滚的内存状态
	type rollbackInfo struct {
		moneyID uint32
		amount  int64
		itemID  uint32
		count   uint32
	}
	var rollbacks []rollbackInfo
	var bagSnapshot map[uint32]*protocol.ItemSt

	// 在开始操作前记录背包快照（用于回滚）
	if bagSys != nil {
		bagSnapshot = bagSys.GetItemsSnapshot()
	}

	// 执行所有消耗操作（不再需要数据库事务，数据已存储在BinaryData中）
	for _, item := range normalized {
		switch item.ItemType {
		case uint32(protocol.ItemType_ItemTypeMoney):
			if moneySys == nil {
				// 回滚已操作的内存状态
				if bagSys != nil && bagSnapshot != nil {
					bagSys.RestoreItemsSnapshot(bagSnapshot)
				}
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money system not ready")
			}
			currentAmount := moneySys.GetAmount(item.ItemId)
			rollbacks = append(rollbacks, rollbackInfo{moneyID: item.ItemId, amount: currentAmount})
			// 使用系统方法更新余额
			if err := moneySys.UpdateBalanceTx(ctx, item.ItemId, -int64(item.Count), nil); err != nil {
				// 回滚已操作的内存状态
				for _, rb := range rollbacks {
					if rb.moneyID > 0 && moneySys != nil {
						moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
					}
				}
				if bagSys != nil && bagSnapshot != nil {
					bagSys.RestoreItemsSnapshot(bagSnapshot)
				}
				return err
			}
		default:
			if bagSys == nil {
				// 回滚已操作的内存状态
				for _, rb := range rollbacks {
					if rb.moneyID > 0 && moneySys != nil {
						moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
					}
				}
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag system not ready")
			}
			// 先检查是否有足够物品
			if !bagSys.HasItem(item.ItemId, uint32(item.Count)) {
				// 回滚已操作的内存状态
				for _, rb := range rollbacks {
					if rb.moneyID > 0 && moneySys != nil {
						moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
					}
				}
				if bagSnapshot != nil {
					bagSys.RestoreItemsSnapshot(bagSnapshot)
				}
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
			}

			// 更新内存状态
			if err := bagSys.RemoveItemTx(item.ItemId, uint32(item.Count)); err != nil {
				// 回滚已操作的内存状态
				for _, rb := range rollbacks {
					if rb.moneyID > 0 && moneySys != nil {
						moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
					}
				}
				if bagSnapshot != nil {
					bagSys.RestoreItemsSnapshot(bagSnapshot)
				}
				return err
			}

			// 发布事件
			pr.Publish(gevent.OnItemRemove, map[string]interface{}{
				"item_id": item.ItemId,
				"count":   item.Count,
			})
		}
	}

	return nil
}

func (pr *PlayerRole) GrantRewards(ctx context.Context, items []*jsonconf.ItemAmount) error {
	normalized := normalizeAmounts(items)
	if len(normalized) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = pr.WithContext(nil)
	}
	bagSys := entitysystem.GetBagSys(ctx)
	moneySys := entitysystem.GetMoneySys(ctx)

	// 记录需要回滚的内存状态
	type rollbackInfo struct {
		moneyID  uint32
		amount   int64
		itemID   uint32
		key      uint32
		count    uint32                      // 用于背包物品回滚
		snapshot map[uint32]*protocol.ItemSt // 用于背包快照回滚
	}
	var rollbacks []rollbackInfo
	var bagSnapshot map[uint32]*protocol.ItemSt

	// 执行所有奖励操作（不再需要数据库事务，数据已存储在BinaryData中）
	for _, item := range normalized {
		switch item.ItemType {
		case uint32(protocol.ItemType_ItemTypeMoney):
			if moneySys == nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "money system not ready")
			}
			currentAmount := moneySys.GetAmount(item.ItemId)
			rollbacks = append(rollbacks, rollbackInfo{moneyID: item.ItemId, amount: currentAmount})
			// 使用AddMoney方法，会自动路由到特殊系统（如经验到level_sys）
			if err := moneySys.AddMoney(ctx, item.ItemId, int64(item.Count)); err != nil {
				// 回滚已操作的内存状态
				for _, rb := range rollbacks {
					if rb.moneyID > 0 {
						moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
					}
				}
				if bagSys != nil && bagSnapshot != nil {
					bagSys.RestoreItemsSnapshot(bagSnapshot)
				}
				return err
			}
		default:
			if bagSys == nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag system not ready")
			}
			// 记录快照用于回滚（只记录一次）
			if bagSnapshot == nil {
				bagSnapshot = bagSys.GetItemsSnapshot()
			}
			// 更新内存状态
			if err := bagSys.AddItemTx(item.ItemId, uint32(item.Count), item.Bind); err != nil {
				// 检查是否是背包满的错误
				if err.Error() == "bag is full" || strings.Contains(err.Error(), "bag is full") {
					// 背包满，通过邮件发放
					mailSys := entitysystem.GetMailSys(ctx)
					if mailSys != nil {
						// 构建奖励列表（转换为ItemSt格式）
						rewards := []*jsonconf.ItemSt{
							{
								ItemId: item.ItemId,
								Count:  uint32(item.Count),
								Type:   item.ItemType,
							},
						}
						// 发送邮件
						mailTitle := "背包奖励"
						mailContent := fmt.Sprintf("由于背包空间不足，以下奖励已通过邮件发放：物品ID %d x %d", item.ItemId, item.Count)
						if err := mailSys.SendCustomMail(ctx, mailTitle, mailContent, rewards); err != nil {
							log.Errorf("SendCustomMail failed: %v", err)
							// 邮件发送失败，回滚并返回错误
							for _, rb := range rollbacks {
								if rb.moneyID > 0 {
									moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
								}
							}
							if bagSnapshot != nil {
								bagSys.RestoreItemsSnapshot(bagSnapshot)
							}
							return customerr.Wrap(err)
						}
						// 邮件发送成功，继续处理下一个奖励
						continue
					} else {
						// 邮件系统不可用，回滚并返回错误
						for _, rb := range rollbacks {
							if rb.moneyID > 0 {
								moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
							}
						}
						if bagSnapshot != nil {
							bagSys.RestoreItemsSnapshot(bagSnapshot)
						}
						return customerr.Wrap(err)
					}
				}
				// 其他错误，回滚并返回
				for _, rb := range rollbacks {
					if rb.moneyID > 0 {
						moneySys.UpdateBalanceOnlyMemory(rb.moneyID, rb.amount)
					}
				}
				if bagSnapshot != nil {
					bagSys.RestoreItemsSnapshot(bagSnapshot)
				}
				return err
			}

			// 发布事件
			pr.Publish(gevent.OnItemAdd, map[string]interface{}{
				"item_id": item.ItemId,
				"count":   item.Count,
			})
		}
	}

	return nil
}

func normalizeAmounts(items []*jsonconf.ItemAmount) []*jsonconf.ItemAmount {
	result := make([]*jsonconf.ItemAmount, 0, len(items))
	for _, item := range items {
		if item == nil || item.Count <= 0 {
			continue
		}
		cp := item.Clone()
		if cp == nil {
			continue
		}
		result = append(result, cp)
	}
	return result
}

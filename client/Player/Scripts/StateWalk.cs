using Godot;

#nullable enable

namespace PostApocGame.GameLogic;

/// <summary>
/// 玩家“行走”状态：根据输入移动并播放 walk_* 动画
/// </summary>
public partial class StateWalk : State
{
	public override void Enter()
	{
		if (!HasPlayer)
		{
			GD.PrintErr("[StateWalk] Enter: 未绑定 Player，无法执行");
			return;
		}

		// 进入 Walk 时立即刷新一次动画
		Player.UpdateAnimation("walk");
	}

	public override State? Process(float delta)
	{
		if (!HasPlayer)
		{
			return null;
		}

		// 根据输入更新方向与状态
		RefreshMovementAndAnimation("walk");

		// 如果不再移动（State 已经变成 idle），则切换回 IdleState
		bool isMoving = Player.State == "move";
		if (!isMoving)
		{
			State? idleState = FindSiblingState<StateIdle>();
			if (idleState == null)
			{
				GD.PrintErr("[StateWalk] Process: 未找到 StateIdle 状态节点");
			}
			return idleState;
		}

		return null;
	}

	public override State? Physics(float delta)
	{
		if (!HasPlayer)
		{
			return null;
		}

		// 行走状态下根据当前方向和 MoveSpeed 设置速度，再驱动物理移动
		Vector2 direction = Player.Direction;
		Vector2 velocity = direction * Player.MoveSpeed;
		Player.Velocity = velocity;
		Player.MoveAndSlide();
		return null;
	}

	public override State? HandleInput(InputEvent @event)
	{
		if (!HasPlayer)
		{
			return null;
		}

		return TryEnterAttack(@event);
	}
}



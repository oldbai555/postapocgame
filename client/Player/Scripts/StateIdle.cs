using Godot;

#nullable enable

namespace PostApocGame.GameLogic;

/// <summary>
/// 玩家“待机”状态：站立不动，等待输入触发行走
/// </summary>
public partial class StateIdle : State
{
	public override void Enter()
	{
		if (!HasPlayer)
		{
			GD.PrintErr("[StateIdle] Enter: 未绑定 Player，无法执行");
			return;
		}

		// 进入 Idle 时同步一次动画（例如从其他状态切回）
		Player.UpdateAnimation("idle");
	}

	public override State? Process(float delta)
	{
		if (!HasPlayer)
		{
			return null;
		}

		// 读取输入并更新方向 / 状态 / 动画
		RefreshMovementAndAnimation("idle");

		// 如果已经变成移动状态，则切换到 WalkState
		bool isMoving = Player.State == "move";
		if (isMoving)
		{
			// 查找同一状态机下的 WalkState 实例
			State? walkState = FindSiblingState<StateWalk>();
			if (walkState == null)
			{
				GD.PrintErr("[StateIdle] Process: 未找到 StateWalk 状态节点");
			}
			return walkState;
		}

		return null;
	}

	public override State? Physics(float delta)
	{
		if (!HasPlayer)
		{
			return null;
		}

		// Idle 状态下速度应为 0，只是让物理系统跑一遍保持正常
		Player.Velocity = Vector2.Zero;
		Player.MoveAndSlide();
		return null;
	}

	public override State? HandleInput(InputEvent @event)
	{
		if (!HasPlayer)
		{
			return null;
		}

		State? attackState = TryEnterAttack(@event);
		return attackState;
	}
}

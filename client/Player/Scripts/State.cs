using Godot;
using System;

#nullable enable

namespace PostApocGame.GameLogic;

/// <summary>
/// 玩家状态基类：Idle / Walk 等都继承自这里。
/// </summary>
public abstract partial class State : Node
{
	private Player? _player;

	/// <summary>
	/// 获取当前状态绑定的 Player，若未绑定则抛出异常，方便尽早发现配置问题。
	/// </summary>
	protected Player Player => _player ?? throw new InvalidOperationException("State has not been attached to a Player.");

	/// <summary>
	/// 当前状态是否已经绑定 Player。
	/// </summary>
	protected bool HasPlayer => _player != null;

	internal void AttachPlayer(Player player)
	{
		_player = player;
	}

	public virtual void Enter()
	{
	}

	public virtual void Exit()
	{
	}

	public virtual State? Process(float delta)
	{
		return null;
	}

	public virtual State? Physics(float delta)
	{
		return null;
	}

	public virtual State? HandleInput(InputEvent @event)
	{
		return null;
	}

	/// <summary>
	/// 工具函数：在同级节点中查找指定的状态类型。
	/// </summary>
	protected T? FindSiblingState<T>() where T : State
	{
		Node? parent = GetParent();
		if (parent == null)
		{
			return null;
		}

		foreach (Node child in parent.GetChildren())
		{
			if (child is T typed)
			{
				return typed;
			}
		}

		return null;
	}

	/// <summary>
	/// 刷新输入方向/状态，并在需要时播放对应动画。返回本次状态是否发生变化。
	/// </summary>
	protected bool RefreshMovementAndAnimation(string animationPrefix)
	{
		if (!HasPlayer)
		{
			return false;
		}

		bool directionChanged = Player.SetDirection();
		bool stateChanged = Player.SetState();

		if (directionChanged || stateChanged)
		{
			Player.UpdateAnimation(animationPrefix);
		}

		return stateChanged;
	}

	/// <summary>
	/// 公共攻击输入检测，命中后返回攻击状态实例。
	/// </summary>
	protected State? TryEnterAttack(InputEvent @event)
	{
		if (!HasPlayer)
		{
			return null;
		}

		bool isAttackKey = @event is InputEventKey key && key.Pressed && key.PhysicalKeycode == Key.Z;
		bool isAttackAction = InputMap.HasAction("attack") && @event.IsActionPressed("attack");
		if (!isAttackKey && !isAttackAction)
		{
			return null;
		}

		State? attackState = FindSiblingState<StateAttack>();
		if (attackState == null)
		{
			GD.PrintErr("[State] TryEnterAttack: 未找到 StateAttack 状态节点");
		}
		return attackState;
	}
}
 
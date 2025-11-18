using Godot;
using PostApocGame.GameLogic;

#nullable enable

namespace PostApocGame.GameLogic;

/// <summary>
/// 玩家状态机，设计类比 Go 服务端的“状态调度器”：负责当前 State 的生命周期与切换。
/// </summary>
public partial class Player_state_machine : Node
{
	private readonly Godot.Collections.Array<State> _states = new();
	private State? _currentState;

	public override void _Ready()
	{
		ProcessMode = ProcessModeEnum.Disabled;
	}

	public override void _Process(double delta)
	{
		if (_currentState == null)
		{
			GD.PrintErr("[PlayerStateMachine] _Process 在没有当前状态的情况下被调用，请确认已初始化。");
			return;
		}

		State? next = _currentState.Process((float)delta);
		if (next != null && next != _currentState)
		{
			ChangeState(next);
		}
	}

	public override void _PhysicsProcess(double delta)
	{
		if (_currentState == null)
		{
			GD.PrintErr("[PlayerStateMachine] _PhysicsProcess 在没有当前状态的情况下被调用，请确认已初始化。");
			return;
		}

		State? next = _currentState.Physics((float)delta);
		if (next != null && next != _currentState)
		{
			ChangeState(next);
		}
	}

	public override void _UnhandledInput(InputEvent @event)
	{
		HandleInput(@event);
	}

	/// <summary>
	/// 处理输入事件（由 Player 调用），转发给当前状态。
	/// </summary>
	public void HandleInput(InputEvent @event)
	{
		if (_currentState == null)
		{
			return;
		}

		// 输入事件只转发给当前 State，本类不做任何业务判断
		State? next = _currentState.HandleInput(@event);
		if (next != null && next != _currentState)
		{
			ChangeState(next);
		}
	}

	/// <summary>
	/// 初始化状态机（由 Player 调用），会把 Player 注入所有 State 并切换到首个 State。
	/// </summary>
	public void Initialize(Player player)
	{
		_states.Clear();

		foreach (Node child in GetChildren())
		{
			if (child is State state)
			{
				state.AttachPlayer(player);
				_states.Add(state);
			}
			// 忽略非 State 子节点
		}

		if (_states.Count > 0)
		{
			ChangeState(_states[0]);
			ProcessMode = ProcessModeEnum.Inherit;
		}
		else
		{
			GD.PrintErr("[PlayerStateMachine] 未找到任何状态节点！");
		}
	}

	// 对外切换状态的统一入口
	public void ChangeState(State newState)
	{
		if (newState == null || newState == _currentState)
		{
			if (newState == null)
			{
				GD.PrintErr("[PlayerStateMachine] ChangeState: 尝试切换到 null 状态，已忽略");
			}
			return;
		}

		if (_currentState != null)
		{
			_currentState.Exit();
		}

		_currentState = newState;

		_currentState.Enter();
	}
}
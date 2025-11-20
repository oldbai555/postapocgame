using Godot;
using System;

#nullable enable

namespace PostApocGame.GameLogic;

[GodotClassName("Player")]
public partial class Player : CharacterBody2D
{
	private const float DirectionEpsilon = 0.0001f;
	private const string MoveLeftAction = "move_left";
	private const string MoveRightAction = "move_right";
	private const string MoveUpAction = "move_up";
	private const string MoveDownAction = "move_down";

	[Export]
	public float MoveSpeed { get; set; } = 100f;

	private AnimationPlayer? _animationPlayer;
	private Sprite2D? _sprite;
	[Export]
	public NodePath? StateMachinePath { get; set; }

	private Player_state_machine? _stateMachine;
	private Vector2 _cardinalDirection = Vector2.Zero;
	private Vector2 _direction = Vector2.Zero;
	private string _state = "idle";

	public Vector2 CardinalDirection => _cardinalDirection;
	public Vector2 Direction => _direction;
	public string State => _state;
	public bool IsFacingLeft => _sprite != null && _sprite.Scale.X < 0f;

	[Signal]
	public delegate void DirectionChangedEventHandler(Vector2 newDirection);

	public override void _Ready()
	{
		_animationPlayer = GetNodeOrNull<AnimationPlayer>("AnimationPlayer");
		_sprite = GetNodeOrNull<Sprite2D>("Sprite2D");

		_stateMachine = ResolveStateMachine();
		if (_stateMachine == null)
		{
			GD.PrintErr("[Player] 未找到 Player_state_machine，输入/状态逻辑不会执行。请检查节点路径或 StateMachinePath 配置。");
		}
		else
		{
			GD.Print("[Player] 初始化 Player_state_machine 成功。");
			_stateMachine.Initialize(this);
		}
	}

	public override void _Process(double delta)
	{
	}

	public override void _PhysicsProcess(double delta)
	{
	}

	public override void _UnhandledInput(InputEvent @event)
	{
		_stateMachine?.HandleInput(@event);
	}

	private Player_state_machine? ResolveStateMachine()
	{
		if (StateMachinePath != null && !StateMachinePath.IsEmpty)
		{
			return GetNodeOrNull<Player_state_machine>(StateMachinePath);
		}

		return GetNodeOrNull<Player_state_machine>("StateMachine");
	}

	/// <summary>
	/// 处理输入并更新运动方向和主方向。
	/// 设计类比服务端：状态机（handler）负责调用该方法获取“本帧输入”，再决定是否写 Velocity。
	/// </summary>
	public bool SetDirection()
	{
		Vector2 inputDirection = Input.GetVector(MoveLeftAction, MoveRightAction, MoveUpAction, MoveDownAction);
		Vector2 oldDirection = _direction;

		_direction = inputDirection == Vector2.Zero ? Vector2.Zero : inputDirection.Normalized();
		if (_direction == Vector2.Zero)
		{
			return false;
		}

		bool hasInput = inputDirection.LengthSquared() > DirectionEpsilon;
		if (hasInput)
		{
			bool isHorizontal = Mathf.Abs(inputDirection.X) > Mathf.Abs(inputDirection.Y);
			if (isHorizontal)
			{
				_cardinalDirection = inputDirection.X > 0 ? Vector2.Right : Vector2.Left;
			}
			else
			{
				_cardinalDirection = inputDirection.Y > 0 ? Vector2.Down : Vector2.Up;
			}
		}

		if (oldDirection == _direction)
		{
			return false;
		}
		EmitSignal(SignalName.DirectionChanged, _direction);
		UpdateFacing();

		return true;
	}

	/// <summary>
	/// 根据当前方向设置角色状态（idle / move）。
	/// 注意：这里的 state 仅用于表现层（动画），与服务端战斗状态解耦。
	/// </summary>
	public bool SetState()
	{
		// 记录旧状态
		string oldState = _state;

		bool isMoving = _direction.LengthSquared() > DirectionEpsilon;
		_state = isMoving ? "move" : "idle";

		return !_state.Equals(oldState, StringComparison.Ordinal);
	}

	/// <summary>
	/// 获取当前方向对应的动画方向名称（down / up / side）
	/// </summary>
	public string AnimDirection()
	{
		bool isHorizontal = Mathf.Abs(_cardinalDirection.X) >= Mathf.Abs(_cardinalDirection.Y);
		return isHorizontal
			? "side"
			: (_cardinalDirection.Y < 0 ? "up" : "down");
	}

	/// <summary>
	/// 根据当前方向，选择并播放对应前缀的动画（前缀由状态机传入）。
	/// - 状态机决定业务语义（idle / walk / attack ...）
	/// - Player 负责把主方向映射为 down/up/side，拼出最终动画名
	/// </summary>
	public void UpdateAnimation(string prefix)
	{
		if (_animationPlayer == null)
		{
			return;
		}

		string animName = $"{prefix}_{AnimDirection()}";
		if (!_animationPlayer.HasAnimation(animName))
		{
			return;
		}

		bool needPlay = _animationPlayer.CurrentAnimation != animName || !_animationPlayer.IsPlaying();
		if (needPlay)
		{
			_animationPlayer.Play(animName);
		}
	}

	private void UpdateFacing()
	{
		if (_sprite == null)
		{
			return;
		}
		Vector2 scale = _sprite.Scale;
		scale.X = _cardinalDirection == Vector2.Left ? -1f : 1f;
		_sprite.Scale = scale;
	}
}

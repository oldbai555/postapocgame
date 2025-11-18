using Godot;

#nullable enable

namespace PostApocGame.GameLogic;

/// <summary>
/// 玩家"攻击"状态：播放攻击动画，攻击完成后根据输入切换到 Idle 或 Walk 状态
/// </summary>
public partial class StateAttack : State
{
	private AnimationPlayer? _animationPlayer;
	private AnimationPlayer? _attackEffectAnimationPlayer;
	private Sprite2D? _attackEffectSprite;
	private AudioStreamPlayer2D? _attackAudioPlayer;
	private bool _animationDone;

	[Export]
	public AudioStream? AttackSound { get; set; }

	[Export(PropertyHint.Range, "1,20,0.5")]
	public float DecelerateSpeed { get; set; } = 5f;

	public override void _Ready()
	{
		Node? parent = GetParent();
		if (parent == null)
		{
			return;
		}
		Node? playerNode = parent.GetParent();
		if (playerNode == null)
		{
			return;
		}
		_animationPlayer = playerNode.GetNodeOrNull<AnimationPlayer>("AnimationPlayer");
		_attackEffectAnimationPlayer = playerNode.GetNodeOrNull<AnimationPlayer>("Sprite2D/AttackEffectSprite/AttackEffectAnimationPlayer");
		_attackEffectSprite = playerNode.GetNodeOrNull<Sprite2D>("Sprite2D/AttackEffectSprite");
		_attackAudioPlayer = playerNode.GetNodeOrNull<AudioStreamPlayer2D>("Audio/AudioStreamPlayer2D");

		if (_animationPlayer != null)
		{
			_animationPlayer.AnimationFinished += OnAnimationFinished;
		}
	}

	public override void Enter()
	{
		if (!HasPlayer)
		{
			GD.PrintErr("[StateAttack] Enter: player 未绑定");
			return;
		}

		_animationDone = false;
		Player.UpdateAnimation("attack");

		PlayAttackEffect();
		PlayAttackSound();
	}

	public override void Exit()
	{
		_animationDone = false;
	}

	public override State? Process(float delta)
	{
		if (!HasPlayer)
		{
			return null;
		}

		Player.Velocity -= Player.Velocity * delta * DecelerateSpeed;

		if (!_animationDone)
		{
			return null;
		}

		Player.SetDirection();
		Player.SetState();
		bool isMoving = Player.State == "move";
		return isMoving ? FindSiblingState<StateWalk>() : FindSiblingState<StateIdle>();
	}

	public override State? Physics(float delta)
	{
		if (!HasPlayer)
		{
			return null;
		}

		Player.MoveAndSlide();
		return null;
	}

	/// <summary>
	/// 动画完成回调：当攻击动画播放完成时调用
	/// </summary>
	private void OnAnimationFinished(StringName animName)
	{
		var name = animName.ToString();
		if (name.StartsWith("attack"))
		{
			_animationDone = true;
		}
	}


	/// <summary>
	/// 播放攻击音效（包含可选的随机音调偏移）
	/// </summary>
	private void PlayAttackSound()
	{
		if (_attackAudioPlayer == null)
		{
			return;
		}

		if (AttackSound != null)
		{
			_attackAudioPlayer.Stream = AttackSound;
		}

		_attackAudioPlayer.PitchScale = 1.0f;
		_attackAudioPlayer.Play();
	}

	private void PlayAttackEffect()
	{
		if (_attackEffectAnimationPlayer == null || _attackEffectSprite == null || !HasPlayer)
		{
			return;
		}

		string animName = $"attack_{Player.AnimDirection()}";
		if (!_attackEffectAnimationPlayer.HasAnimation(animName))
		{
			return;
		}

		_attackEffectAnimationPlayer.Play(animName);
	}

	public override void _ExitTree()
	{
		if (_animationPlayer != null)
		{
			_animationPlayer.AnimationFinished -= OnAnimationFinished;
		}
		base._ExitTree();
	}
}

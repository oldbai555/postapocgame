using Godot;

#nullable enable

namespace PostApocGame.GameLogic;

public partial class PlayerInteractionsHost : Node2D
{
	[Export]
	public NodePath? PlayerPath { get; set; }

	private Player? _player;

	public override void _Ready()
	{
		_player = ResolvePlayer();
		if (_player == null)
		{
			GD.PrintErr("[PlayerInteractionsHost] 未找到 Player 节点，方向同步失效。");
			return;
		}

		_player.DirectionChanged += OnDirectionChanged;
	}

	private Player? ResolvePlayer()
	{
		if (PlayerPath != null && !PlayerPath.IsEmpty)
		{
			return GetNodeOrNull<Player>(PlayerPath);
		}

		return GetParent() as Player ?? GetParent()?.GetParent() as Player;
	}

	private void OnDirectionChanged(Vector2 newDirection)
	{
		float degrees = 0f;
		if (newDirection == Vector2.Up)
		{
			degrees = 180f;
		}
		else if (newDirection == Vector2.Left)
		{
			degrees = 90f;
		}
		else if (newDirection == Vector2.Right)
		{
			degrees = -90f;
		}

		RotationDegrees = degrees;
	}

	public override void _ExitTree()
	{
		if (_player != null)
		{
			_player.DirectionChanged -= OnDirectionChanged;
		}

		base._ExitTree();
	}
}

using Godot;

public partial class Plant : Node2D
{
	public override void _Ready()
	{
		var hitBox = GetNode<HitBox>("Hitbox");
		hitBox.Damaged += TakeDamage;
	}

	public void TakeDamage(int damage)
	{
		QueueFree();
	}
}

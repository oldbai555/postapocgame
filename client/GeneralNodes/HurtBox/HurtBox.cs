using Godot;

public partial class HurtBox : Area2D
{
	[Export]
	public int Damage { get; set; } = 1;

	public override void _Ready()
	{
		AreaEntered += OnAreaEntered;
	}

	public override void _Process(double delta)
	{
	}

	private void OnAreaEntered(Area2D area)
	{
		if (area is HitBox hitBox)
		{
			hitBox.TakeDamage(Damage);
		}
	}
}

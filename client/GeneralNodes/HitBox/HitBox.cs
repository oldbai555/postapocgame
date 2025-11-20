using Godot;

public partial class HitBox : Area2D
{
	[Signal]
	public delegate void DamagedEventHandler(int damage);

	public override void _Ready()
	{
	}

	public override void _Process(double delta)
	{
	}

	public void TakeDamage(int damage)
	{
		GD.Print($"TakeDamage:{damage}");
		EmitSignal(SignalName.Damaged, damage);
	}
}

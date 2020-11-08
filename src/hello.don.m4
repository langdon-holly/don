define(half_adder, merge (@
	(@
		0:
		0:
	@) split :0 :0
	(@
		1:
		0:
	@) split merge (@
		:1 :0
		:0 :1
	)
	(@
		0:
		1:
	@) split :1 :1
) prod)

define(full_adder, merge (@
	(@
		0:
		0:
	@) split :0 :0 :0
	(@
		1:
		0:
	@) split merge (@
		:1 :0 :0
		:0 :1 :0
		:0 :0 :1
	)
	(@
		0:
		1:
	@) split merge (@
		:1 :1 :0
		:1 :0 :1
		:0 :1 :1
	)
	(@
		1:
		1:
	@) split :1 :1 :1
) prod)

#(
	[+1] append ride!(
		full_adder (@
			:a :0
			:b :0
			:1
		)
	) (@
		:ab
		:carry
	)
	:out
) (
	ab: [-1] over (
		:a:
		:b:
	)
	(
		out: 0:
		carry:
	@) half_adder (@
		:0 :a
		:0 :b
	)
)

(
	map!:out
	8:!:2
) (
	(
		7: out:
		8: 2:
	@) full_adder :7
	I
) (
	(
		6: out:
		7: 2:
	@) full_adder :6
	I
) (
	(
		5: out:
		6: 2:
	@) full_adder :5
	I
) (
	(
		4: out:
		5: 2:
	@) full_adder :4
	I
) (
	(
		3: out:
		4: 2:
	@) full_adder :3
	I
) (
	(
		2: out:
		3: 2:
	@) full_adder :2
	I
) (
	(
		1: out:
		2: 2:
	@) full_adder :1
	I
) (
	(
		0: out:
		1: 2:
	@) half_adder :0
	I
) (
	map!0:
	map!1:
@)

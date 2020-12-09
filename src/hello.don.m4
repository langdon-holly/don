define(half_adder, <- (
	(
		0:
		0:
	) => :0 :0
	(
		1:
		0:
	) => <- (
		:1 :0
		:0 :1
	) >
	(
		0:
		1:
	) => :1 :1
) > prod)

define(half_adder_from_the_future, <- (
	(
		0:
		0:
	) => <= (
		:0
		:0
	)
	(
		1:
		0:
	) => <- (
		<= (
			:0
			:1
		)
		<= (
			:1
			:0
		)
	) ->
	(
		0:
		1:
	) => <= (
		:1
		:1
	)
) ->)

define(full_adder, <- (
	(
		0:
		0:
	) => :0 :0 :0
	(
		1:
		0:
	) => <- (
		:1 :0 :0
		:0 :1 :0
		:0 :0 :1
	) >
	(
		0:
		1:
	) => <- (
		:1 :1 :0
		:1 :0 :1
		:0 :1 :1
	) >
	(
		1:
		1:
	) => :1 :1 :1
) > prod)

< (
	map!:out
	8:!:2
) > < (
	< (
		7: out:
		8: 2:
	) full_adder :7
	I
) > < (
	< (
		6: out:
		7: 2:
	) full_adder :6
	I
) > < (
	< (
		5: out:
		6: 2:
	) full_adder :5
	I
) > < (
	< (
		4: out:
		5: 2:
	) full_adder :4
	I
) > < (
	< (
		3: out:
		4: 2:
	) full_adder :3
	I
) > < (
	< (
		2: out:
		3: 2:
	) full_adder :2
	I
) > < (
	< (
		1: out:
		2: 2:
	) full_adder :1
	I
) > < (
	< (
		0: out:
		1: 2:
	) half_adder :0
	I
) > < (
	map!0:
	map!1:
)

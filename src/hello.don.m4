define(wait, :out rec!(
	(
		out: :T:
		F?: :F
	) yet (
		_: :T :_
		?: :T?
	)
	(
		out: F: :T
		T?: :F
	) yet (
		_: :F :_
		?: :F?
	)
) (
	T?: :for
	:_:
))

define(half_adder, (
	(@
		0:
		0:
	) :0 :0
	(@
		1:
		0:
	) (
		:1 :0
		:0 :1
	)
	(@
		0:
		1:
	) :1 :1
) prod)

define(full_adder, (
	(@
		0:
		0:
	) :0 :0 :0
	(@
		1:
		0:
	) (
		:1 :0 :0
		:0 :1 :0
		:0 :0 :1
	)
	(@
		0:
		1:
	) (
		:1 :1 :0
		:1 :0 :1
		:0 :1 :1
	)
	(@
		1:
		1:
	) :1 :1 :1
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
	(
		7:
		#carry to output:
		8:
	@) full_adder (@
		:7 :a
		:7 :b
		:7 :carry
	)
	:out
) (
	(
		out: 6:
		carry: 7:
	@) full_adder (@
		:6 :a
		:6 :b
		:6 :carry
	)
	:out:
	:a:
	:b:
) (
	(
		out: 5:
		carry: 6:
	@) full_adder (@
		:5 :a
		:5 :b
		:5 :carry
	)
	:out:
	:a:
	:b:
) (
	(
		out: 4:
		carry: 5:
	@) full_adder (@
		:4 :a
		:4 :b
		:4 :carry
	)
	:out:
	:a:
	:b:
) (
	(
		out: 3:
		carry: 4:
	@) full_adder (@
		:3 :a
		:3 :b
		:3 :carry
	)
	:out:
	:a:
	:b:
) (
	(
		out: 2:
		carry: 3:
	@) full_adder (@
		:2 :a
		:2 :b
		:2 :carry
	)
	:out:
	:a:
	:b:
) (
	(
		out: 1:
		carry: 2:
	@) full_adder (@
		:1 :a
		:1 :b
		:1 :carry
	)
	:out:
	:a:
	:b:
) (
	(
		out: 0:
		carry: 1:
	@) half_adder (@
		:0 :a
		:0 :b
	)
	:a:
	:b:
) (
	a:
	b:
@)

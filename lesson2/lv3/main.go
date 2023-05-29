package main

import "fmt"

type Person struct {
	name  string
	hp    int
	atk   int
	exp   int
	level int
}

type Attacker interface {
	attack(*Person)
}

func (p *Person) attack(target *Person) {
	target.hp -= p.atk
	fmt.Println(p.name, "attacked", target.name)
	fmt.Println(target.name, "got", p.atk, "damage")
}

func main() {
	p1 := &Person{
		name:  "ding_zhen",
		hp:    100,
		atk:   10,
		exp:   0,
		level: 1,
	}

	p2 := &Person{
		name:  "sun_xiao_chuan",
		hp:    114514,
		atk:   10,
		exp:   0,
		level: 1,
	}

	p1.attack(p2)

}

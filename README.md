# krools

Simple yet powerful rules engine for software engineers.

krools execute rules against your data to inference new data and/or guide a
process. It is of passive and forward chaining type.

## Docs for v2

Doc is still progress. But to use engine module you need to import `github.com/krocos/krools/v2` path.

## Docs for v1

## Knowledge base

Knowledge base contains rules you will apply against your data.

Rules grouped to units (MAIN by default).

You need to create a knowledge base once at the start and use it later any
times.

```go
package main

import (
	"context"

	"github.com/krocos/krools"
)

type fireContext struct{}

func main() {
	ctx := context.Background()

	var rule *krools.Rule[*fireContext]

	k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
		Add(rule)

	c := new(fireContext)

	_ = k.FireAllRules(ctx, c)
}

```

## Fire context

It can be any type of data. Rules will be evaluated and executed against it. You
can use it like you want.

When engine works you have access to the fire context from conditions and
actions to examine data passed from outside and create data from actions.

### Container (just the experiment)

You can use `Container` to automate the process of loading valuable variables from context. You can embed container into
your fire context or create a field within of `Container` type. Container have a few methods:
- `Insert` – it inserts any number of arguments into container. Arguments must be of struct type or a pointer to struct
  type, or a pointer to a pointer. If any argument is no struct method panics.
- `Retract` – finds all passed variables type path and drop them from a container.
- `FillIn` – fills in passed variables with respective values from container. All passed values must be a pointers to
  instantiated structs. Method returns `bool` sign of that are all passed variables was filled... Method panics if one of
  args is not a pointer to struct or a pointer to a pointer to a struct and so on.

So, it automates receiving a few sign-variable for rules and to retrieve variables after the engine done.

## Rule

Rule consists from three parts: name, condition (when part), and action (then
part).

When a condition returns true, then krools execute an action.  

Action and condition are something that implements corresponding interfaces

```go
package readme

import (
	"context"
)

type (
	Action[T any] interface {
		Execute(ctx context.Context, fireContext T) error
	}
	Condition[T any] interface {
		IsSatisfiedBy(ctx context.Context, fireContext T) (bool, error)
	}
)
```

Create rule implementing interfaces:
```go
package main

import (
	"context"

	"github.com/krocos/krools"
)

type fireContext struct{}

type someCondition struct{}

func newSomeCondition() *someCondition {
	return new(someCondition)
}

func (a *someCondition) IsSatisfiedBy(ctx context.Context, fireContext *fireContext) (bool, error) {
	return true, nil
}

type someAction struct{}

func newSomeAction() *someAction {
	return new(someAction)
}

func (a *someAction) Execute(ctx context.Context, fireContext *fireContext) error {
	return nil
}

func main() {
	// Your rule
	_ = krools.NewRule[*fireContext]("rule name", newSomeCondition(), newSomeAction())
}
```

Also, you can create rules by using special function types `ActionFn`,
`ConditionFn`.

Create rule using function types:
```go
package main

import (
	"context"

	"github.com/krocos/krools"
)

type fireContext struct{}

func main() {
	// Your rule
	_ = krools.NewRule[*fireContext](
		"rule name",
		// When
		krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
			return true, nil
		}),
		// Then
		krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
			return nil
		}),
	)
}
```

Condition may be `nil` and it means that rule action will be executed anyway.
Also, action may be `nil` too. Why? Rule may have some number of attributes that
need to be applied without any condition and consequent action.

## Order of execution

krools evaluate conditions of rules to determine which actions need to be
executed. It is evaluation.

When krools found rules to execute it consequently execute actions of those
rules. Rules executes in order they were added.

## Cycle

When krools has executed all rules it found, this process repeats. krools do the
same actions: evaluate and execute. It does it until there are no rules to
execute.

There is the maximum number of reevaluations. If your application run into
infinite cycle it breaks the execution with error.

By default, it equals 65535.

You can set this number by calling `SetMaxReevaluations()` on a knowledge base.

## Rule auto retraction

When a rule was executed it will be retracted. It means that the rule will be
ignored on next cycle of evaluation.

So, this code
```go
package main

import (
	"context"
	"fmt"

	"github.com/krocos/krools"
)

type fireContext struct{}

func main() {
	ctx := context.Background()

	rule := krools.NewRule[*fireContext](
		"rule name",
		// When
		krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
			return true, nil
		}),
		// Then
		krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
			fmt.Println("rule executed")

			return nil
		}),
	)

	k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
		Add(rule)

	c := new(fireContext)

	if err := k.FireAllRules(ctx, c); err != nil {
		panic(err)
	}
}
```

Returns
```text
rule executed

```

## Unit

All rules added to knowledge base are added to the MAIN unit. But you can add
rule to any other unit you need.

Units executes rules in order they (units) were added. krools take the first
unit and run evaluation cycle. When there are no rules to execute, krools takes
second unit and so on.

## Attributes of rules

Rules have different attributes for different purposes.

### Salience

You can set salience to manipulate the order of execution. Salience is plain
`int` so it can be with minus. Rules with greater salience executes before
others. Others executes in order they added.

```go
package main

import (
	"context"
	"fmt"

	"github.com/krocos/krools"
)

type fireContext struct{}

func main() {
	ctx := context.Background()

	rule1 := krools.NewRule[*fireContext](
		"rule one",
		// When
		krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
			return true, nil
		}),
		// Then
		krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
			fmt.Println("rule 1")

			return nil
		}),
	)

	rule2 := krools.NewRule[*fireContext](
		"rule two",
		// When
		krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
			return true, nil
		}),
		// Then
		krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
			fmt.Println("rule 2")

			return nil
		}),
	)

	k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
		Add(rule1).
		Add(rule2.Salience(1))

	c := new(fireContext)

	if err := k.FireAllRules(ctx, c); err != nil {
		panic(err)
	}
}
```

Returns
```text
rule 2
rule 1

```

### Retraction

Rule is retracted after execution, but it can retract any other rule.

```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	Add(rule1.Retract("rule two")).
	Add(rule2)
```

Returns
```text
rule 1

```

### Insertion

Rule can be inserted again after retraction. Also, it can insert any other rule.

```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	Add(rule1).
	Add(rule2.Insert("rule one"))
```

Returns
```text
rule 1
rule 2
rule 1

```

As we see the rule two insert the rule one, and it wasn't retracted on second cycle.

If you call Insert() without arguments it will insert the rule itself.

### Unit of rule

You can add rule to your unit calling `Unit()` method of rule.

```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	Add(rule1).
	Add(rule2.Unit("other unit").Salience(1_000_000))
```

This example returns
```text
rule 1
rule 2

```

It's because units executes in order they added. In this example we add `rule1`
to the default unit MAIN before the "other unit" unit.

Also, you can add units like this
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	Add(rule1).
	AddUnit("one", rule2, rule3).
	AddUnit("two", rule4, rule5, rule6)
```

### Activation unit

You can add multiple rules to activation unit. If any of rules of activation
unit will be executed so all other rules of this activation unit retracts.

```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	Add(rule1.ActivationUnit("activation unit")).
	Add(rule3).
	Add(rule2.ActivationUnit("activation unit"))
```

Returns
```text
rule 1
rule 3

```

Execution of `rule1` retracts `rule2`, but `rule3` is not in this activation
unit.

### No loop

When you set up your rule to be inserted it will be evaluated (and maybe
executed) on next cycle. But if you don't need this you can use `NoLoop()`
method. It is very helpful when you returns to the unit again. For example when
you implement recursive handling of some items. So rule stay inserted but omit
looping in cycle.

```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	Add(rule1).
	AddUnit("unit", rule2.Insert().NoLoop())
```

## Action stack

You can write an action once and reuse it by stacking it with actions of other
rules. For this particular reason krools have function `NewActionStack`. So you
can add to your rule `ActionStack` instead of just an action.

Just for illustration
```go
package main

import (
	"context"
	"fmt"

	"github.com/krocos/krools"
)

type fireContext struct{}

func main() {
	logAction := krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
		fmt.Println("==> log something from fire context")
		return nil
	})

	ruleOne := krools.NewRule[*fireContext](
		"rule one",
		// When
		nil,
		// Then
		krools.NewActionStack[*fireContext](
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fmt.Println("do the main work of rule one")
				return nil
			}),
			logAction,
		),
	)

	ruleTwo := krools.NewRule[*fireContext](
		"rule two",
		// When
		nil,
		// Then
		krools.NewActionStack[*fireContext](
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fmt.Println("do the main work of rule two")
				return nil
			}),
			logAction,
		),
	)

	k := krools.NewKnowledgeBase[*fireContext]("stacked actions").
		Add(ruleOne).
		Add(ruleTwo)

	c := new(fireContext)

	if err := k.FireAllRules(context.Background(), c); err != nil {
		panic(err)
	}
}
```

This code returns
```text
do the main work of rule one
==> log something from fire context
do the main work of rule two
==> log something from fire context

```

## Flow of execution

Rule have the attributes to be managed in a unit. But also rules can guide the
flow of execution by activating or deactivating units. Also, it can set focus on
units to execute units again.

### Activation and deactivation of units

Any unit can be deactivated when you create a knowledge base. This means all
rules those units contains will be retracted.

```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
    SetDeactivatedUnits("unit1", "unit2").
    AddUnit(krools.UnitMAIN, rule1).
    AddUnit("unit1", rule2).
	AddUnit("unit2", rule3)
```

This results in
```text
rule 1

```

But you can conditionally activate units when rule was executed by calling
`ActivateUnits()` on a rule
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
    SetDeactivatedUnits("unit1", "unit2").
    AddUnit(krools.UnitMAIN, rule1.ActivateUnits("unit1")).
    AddUnit("unit1", rule2).
    AddUnit("unit2", rule3)
```

It results in
```text
rule 1
rule 2

```

Also, you can deactivate units in such the way. For example, you have
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	AddUnit(krools.UnitMAIN, rule1).
	AddUnit("unit1", rule2).
	AddUnit("unit2", rule3)
```

It results in
```text
rule 1
rule 2
rule 3

```

But if you want, for example, to disable `unit1` if `rule1` was executed, you
can do this by calling `DeactivateUnits()` on a rule like this
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	AddUnit(krools.UnitMAIN, rule1.DeactivateUnits("unit1")).
	AddUnit("unit1", rule2).
	AddUnit("unit2", rule3)
```

And this results in
It results in
```text
rule 1
rule 3

```

`ActivateUnits()` and `DeactivateUnits()` methods of rule inserts and retracts
all rules of units which names passed. But if you do not need to activate or
deactivate all rules in those units, you can just call `Insert()` and/or
`Retract()` methods and pass rule names specifically.

### Focus on units

All rules grouped to units. By default, it's MAIN unit. But if you add a few
units they will be executed in order you add them.

For example
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	AddUnit(krools.UnitMAIN, rule1).
	AddUnit("unit1", rule2).
	AddUnit("unit2", rule3)
```

This creates stack of units:
- MAIN
- unit1
- unit2

But what if you need to execute `unit1` after you execute `unit2` again?

You can call `SetFocus()` on a rule and pass unit names to execute. For example
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
    AddUnit(krools.UnitMAIN, rule1).
    AddUnit("unit1", rule2).
    AddUnit("unit2", rule3.SetFocus("unit1", krools.UnitMAIN))
```

And the stack of units will change to:
- unit1
- MAIN
- unit2

And execution of units will be started again from first unit (`unit1`). Of
course, rules those units contains will be executed if they still inserted.

Our previous example returns
```text
rule 1
rule 2
rule 3

```

But we can activate those units to execute rules again using `ActivateUnits()`
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
    AddUnit(krools.UnitMAIN, rule1).
    AddUnit("unit1", rule2).
    AddUnit("unit2", rule3.ActivateUnits("unit1", krools.UnitMAIN).
    SetFocus("unit1", krools.UnitMAIN))

```

`ActivateUnits()` inserts all rules in those units so the result will be
```text
rule 1
rule 2
rule 3
rule 2
rule 1

```

The NoLoop attribute of rule is very helpful in such scenarios.

## Execution filters

When you call `FireAllRules()` method of knowledge base you can optionally pass
filters which can filter rules to be executed.

For example, we can execute only rules that has suffix `one` like this:
```go
k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
	AddUnit(krools.UnitMAIN, rule1).
	AddUnit("unit1", rule2).
	AddUnit("unit2", rule3)
    
c := new(fireContext)

filter := krools.RuleNameEndsWith[*fireContext]("one") // rule name is "rule one"
    
if err := k.FireAllRules(ctx, c, filter); err != nil {
	panic(err)
}
```

This returns
```text
rule 1

```

There is basic collection of filters:
- `RuleNameStartsWith` – run only rules the name starts with
- `RuleNameEndsWith` – run only rules the name ends with
- `RuleNameMatchRegexp` – run only units the name match a regexp
- `RuleNameMustNotContainsAny` – run only rules the name does NOT contain any of passed substrings
- `RuleNameMustContainsAny` – run only rules the name does contain any of passed substrings
- `RunOnlyUnits` – run only rules that belongs to passed units

## Execution dispatching

When you call `FireAllRules()` you dispatch a single execution of krools. But
what if you need to iterate over a number of items or handle stream of some data
items? What if you need to preserve some parts of fire context between
executions of engine (stateful session)?

To be able to handle such cases krools have special interface `Dispatcher`
```go
type Executor[T any] func(ctx context.Context, fireContext T) error

type Dispatcher[T any] interface {
	Dispatch(ctx context.Context, fireContext T, fireAllRules Executor[T]) error
}
```

So the type that implements `Dispatcher` interface will manage execution
itself by calling `fireAllRules` function. This function starts execution of
engine as usual, but now you can handle in what an order or when.

For example, let's imagine that we want to iterate over some slice of items and
preserve results of this work to inference further. Let's say it will be
something like this: iterate over words and pick only those which are without
letter A and then count those words. It will look like this:
```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/krocos/krools"
)

type wordToProcess struct {
	value string
}

type needToCount struct {
	value bool
}

type fireContext struct {
	// Incoming slice of words.
	words []string

	// word is the word we are working on every iteration.
	word          *wordToProcess
	wordsWithoutA []string

	// needToCount is the sign if we need to count words without letter A.
	needToCount *needToCount

	count int
}

type iterationDispatcher struct{}

func (i *iterationDispatcher) Dispatch(
	ctx context.Context,
	fireContext *fireContext,
	fireAllRules krools.Executor[*fireContext],
) error {
	for _, word := range fireContext.words {

		// Here we change only the word to process, but preserve all other data.
		// And then we dispatch next execution.

		fireContext.word = &wordToProcess{value: word}

		if err := fireAllRules(ctx, fireContext); err != nil {
			return err
		}
	}

	// After iteration, we run rules that waits for needToCount directive.

	fireContext.word = nil
	fireContext.needToCount = &needToCount{value: true}

	return fireAllRules(ctx, fireContext)
}

func main() {
	k := krools.NewKnowledgeBase[*fireContext]("iterate over words").
		Add(krools.NewRule[*fireContext](
			"add the word without A",
			// When
			krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
				return fireContext.word != nil && !strings.Contains(fireContext.word.value, "a"), nil
			}),
			// Then
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fireContext.wordsWithoutA = append(fireContext.wordsWithoutA, fireContext.word.value)
				return nil
			}),
		)).
		Add(krools.NewRule[*fireContext](
			"count words without A",
			// When
			krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
				return fireContext.needToCount != nil && fireContext.needToCount.value, nil
			}),
			// Then
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fireContext.count = len(fireContext.wordsWithoutA)
				return nil
			}),
		))

	c := &fireContext{words: []string{"implements", "interface", "order", "example", "imagine"}}

	if err := k.FireAllRules(context.Background(), c, new(iterationDispatcher)); err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("Number of words without A: %d", c.count))
}
```

Take a look at how many rules you need to implement this. It's only two. Of
course, you need dispatcher for this, but the code of it is trivial. You just
replace parts of the fire context and preserve other parts.

This example returns
```text
Number of words without A: 2

```

Also, you can use dispatcher to handle some streams or handle events, for example
messages from queue.

Just for example of this take a look at this dispatcher
```go
type streamDispatcher struct {
	in  <-chan string
	out chan *needProcessing
}

func newStreamDispatcher(in <-chan string) (*streamDispatcher, <-chan *needProcessing) {
	d := &streamDispatcher{
		in:  in,
		out: make(chan *needProcessing),
	}

	return d, d.out
}

func (d *streamDispatcher) Dispatch(ctx context.Context, fireContext *streamContext, fireAllRules krools.Executor[*streamContext]) error {
	defer close(d.out)

	for s := range d.in {
		fireContext.item = &streamItem{value: s}
		fireContext.needSomeProcessing = nil

		if err := fireAllRules(ctx, fireContext); err != nil {
			return err
		}

		if fireContext.needSomeProcessing != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case d.out <- fireContext.needSomeProcessing:
			}
		}
	}

	return nil
}
```

It takes data from one channel, pass it to krools, and pass results on via other
channel. It's event or stream processing.

## Compatibility with gospec

You can use [gospec](https://github.com/krocos/gospec) composite specifications
as conditions for krools.

It allows you to create sophisticated conditions with no headaches.

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/krocos/gospec"

	"github.com/krocos/krools"
)

type fireContext struct {
	word string
}

type wordLongerThanSpec struct {
	gospec.Spec[*fireContext]
	n int
}

func newWordLongerThanSpec(n int) *wordLongerThanSpec {
	s := &wordLongerThanSpec{n: n}
	s.Spec = gospec.New[*fireContext](s)
	return s
}

func (s *wordLongerThanSpec) IsSatisfiedBy(ctx context.Context, candidate *fireContext) (bool, error) {
	return len([]rune(candidate.word)) > s.n, nil
}

func main() {
	ctx := context.Background()

	// Inline specification
	wordContainsLetterA := gospec.NewInline[*fireContext](func(ctx context.Context, candidate *fireContext) (bool, error) {
		return strings.Contains(candidate.word, "a"), nil
	})

	wordLongerThan4Letters := newWordLongerThanSpec(4)

	printLn := func(msg string) krools.Action[*fireContext] {
		return krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
			fmt.Println(msg)
			return nil
		})
	}

	rule1 := krools.NewRule[*fireContext](
		"word contains A and longer than 4 letters",
		// When
		wordContainsLetterA.And(wordLongerThan4Letters),
		// Then
		printLn("word contains A and longer than 4 letters"),
	)

	rule2 := krools.NewRule[*fireContext](
		"word contains A and shorter than 4 letters",
		// When
		wordContainsLetterA.And(wordLongerThan4Letters.Not()),
		// Then
		printLn("word contains A and longer than 4 letters"),
	)

	rule3 := krools.NewRule[*fireContext](
		"word contains NOT A and longer than 4 letters",
		// When
		wordContainsLetterA.Not().And(wordLongerThan4Letters),
		// Then
		printLn("word contains NOT A and longer than 4 letters"),
	)

	rule4 := krools.NewRule[*fireContext](
		"word contains NOT A and shorter than 4 letters",
		// When
		wordContainsLetterA.Not().And(wordLongerThan4Letters.Not()),
		// Then
		printLn("word contains NOT A and shorter than 4 letters"),
	)

	k := krools.NewKnowledgeBase[*fireContext]("knowledge base").
		Add(rule1).
		Add(rule2).
		Add(rule3).
		Add(rule4)

	c := &fireContext{word: "somelongword :))"}

	if err := k.FireAllRules(ctx, c); err != nil {
		panic(err)
	}
}
```

It returns
```text
word contains NOT A and longer than 4 letters

```

## Examples

### Bus pass card

Let's imagine that we are the transport department. We need to produce a bus
pass card for every person. Also, we need to consider age of a person. So that
if a person younger than 18 we need to produce a child buss pass else an adult
bus pass. Also, we need to consider that from time to time the data of a person
may change, and we need to update a bus pass from child to adult one. Also, we
need to send email for parents when we create a child bus pass.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krocos/krools"
)

type person struct {
	name string
	age  int
}

type childBussPassCard struct {
	id          string
	name        string
	activeUntil string
}

type adultBussPassCard struct {
	id          string
	name        string
	activeUntil string
}

type createChildBussPass struct {
	childBussPassCard *childBussPassCard
}

type createAdultBussPass struct {
	adultBussPassCard *adultBussPassCard
}

type deactivateChildBussPass struct {
	childBussPassCard *childBussPassCard
}

type sendMessageForParents struct {
	message           string
	childBussPassCard *childBussPassCard
}

type fireContext struct {
	// Incoming data.

	person *person

	childBussPassCard *childBussPassCard
	adultBussPassCard *adultBussPassCard

	// Directives. Output data.

	createChildBussPass *createChildBussPass
	createAdultBussPass *createAdultBussPass

	deactivateChildBussPass *deactivateChildBussPass

	sendMessageForParents *sendMessageForParents
}

func main() {
	ctx := context.Background()

	contexts := []*fireContext{
		// John already have the child buss pass card. Adult card must be
		// created. The adult bus pass must be deactivated.
		{
			person: &person{
				name: "John",
				age:  19,
			},
			childBussPassCard: &childBussPassCard{
				id:   uuid.New().String(),
				name: "John",
			},
		},
		// Alex have no cards. New child bus pass card must be created. An
		// email for his parents must be sent.
		{
			person: &person{
				name: "Alex",
				age:  16,
			},
		},
		// Peter already have the child bus pass. Nothing to do.
		{
			person: &person{
				name: "Peter",
				age:  17,
			},
			childBussPassCard: &childBussPassCard{
				id:          uuid.New().String(),
				name:        "Peter",
				activeUntil: "2028",
			},
		},
	}

	k := newKnowledgeBase()

	for _, c := range contexts {
		if err := k.FireAllRules(ctx, c); err != nil {
			panic(err)
		}

		fmt.Println(c.person.name)
		interpretResults(c)
	}
}

func newKnowledgeBase() *krools.KnowledgeBase[*fireContext] {
	return krools.NewKnowledgeBase[*fireContext]("control bus pass card of person").
		Add(krools.NewRule[*fireContext](
			"create a child buss pass card if doesn't exists",
			// When
			krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
				return fireContext.person.age <= 18 && fireContext.childBussPassCard == nil, nil
			}),
			// Then
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fireContext.childBussPassCard = &childBussPassCard{
					id:   uuid.New().String(),
					name: fireContext.person.name,
					activeUntil: time.Date(time.Now().Year()+2, 1, 1, 0, 0, 0, 0, time.UTC).
						Format(time.RFC3339),
				}

				fireContext.createChildBussPass = &createChildBussPass{childBussPassCard: fireContext.childBussPassCard}

				return nil
			}),
		)).
		Add(krools.NewRule[*fireContext](
			"create an adult bus pass card if doesn't exists",
			// When
			krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
				return fireContext.person.age > 18 && fireContext.adultBussPassCard == nil, nil
			}),
			// Then
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fireContext.adultBussPassCard = &adultBussPassCard{
					id:   uuid.New().String(),
					name: fireContext.person.name,
					activeUntil: time.Date(time.Now().Year()+2, 1, 1, 0, 0, 0, 0, time.UTC).
						Format(time.RFC3339),
				}

				fireContext.createAdultBussPass = &createAdultBussPass{adultBussPassCard: fireContext.adultBussPassCard}

				return nil
			}),
		)).
		Add(krools.NewRule[*fireContext](
			"deactivate a child bus pass card if the person became adult",
			// When
			krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
				return fireContext.person.age > 18 &&
					fireContext.createAdultBussPass != nil &&
					fireContext.childBussPassCard != nil, nil
			}),
			// Then
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				fireContext.deactivateChildBussPass = &deactivateChildBussPass{childBussPassCard: fireContext.childBussPassCard}

				return nil
			}),
		)).
		Add(krools.NewRule[*fireContext](
			"create message for parents if new child bus pass created",
			// When
			krools.ConditionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) (bool, error) {
				return fireContext.createChildBussPass != nil, nil
			}),
			// Then
			krools.ActionFn[*fireContext](func(ctx context.Context, fireContext *fireContext) error {
				card := fireContext.createChildBussPass.childBussPassCard
				msg := fmt.Sprintf(
					"Hi! %s got new child bus pass card with number '%s' and it active until %s",
					card.name,
					card.id,
					card.activeUntil,
				)

				fireContext.sendMessageForParents = &sendMessageForParents{
					message:           msg,
					childBussPassCard: fireContext.createChildBussPass.childBussPassCard,
				}

				return nil
			}),
		))
}

func interpretResults(c *fireContext) {
	var hasDirectives bool

	if c.createChildBussPass != nil {
		fmt.Println(fmt.Sprintf(
			"New child bus pass card must be created for %s. Active until %s.",
			c.createChildBussPass.childBussPassCard.name,
			c.createChildBussPass.childBussPassCard.activeUntil,
		))

		hasDirectives = true
	}

	if c.createAdultBussPass != nil {
		fmt.Println(fmt.Sprintf(
			"New adult bus pass card must be created for %s. Active until %s.",
			c.createAdultBussPass.adultBussPassCard.name,
			c.createAdultBussPass.adultBussPassCard.activeUntil,
		))

		hasDirectives = true
	}

	if c.deactivateChildBussPass != nil {
		fmt.Println(fmt.Sprintf(
			"The child bus pass card '%s' of %s must be deactivated.",
			c.deactivateChildBussPass.childBussPassCard.id,
			c.deactivateChildBussPass.childBussPassCard.name,
		))

		hasDirectives = true
	}

	if c.sendMessageForParents != nil {
		fmt.Println(fmt.Sprintf(
			"New email for parents of %s must be sent with text: %s",
			c.sendMessageForParents.childBussPassCard.name,
			c.sendMessageForParents.message,
		))

		hasDirectives = true
	}

	if !hasDirectives {
		fmt.Println("Nothing to do.")
	}
}
```

Output
```text
John
        New adult bus pass card must be created for John. Active until
        2025-01-01T00:00:00Z.
        The child bus pass card '6f814151-0377-49aa-9312-935ac9378925' of John must be
        deactivated.
Alex
        New child bus pass card must be created for Alex. Active until
        2025-01-01T00:00:00Z.
        New email for parents of Alex must be sent with text: Hi! Alex got new child bus
        pass card with number '9ee23e55-a2fe-41b9-9820-8aeb455e088a' and it active until
        2025-01-01T00:00:00Z
Peter
        Nothing to do.

```

### Rules reusing

To be able to use conditions & actions in multiple types of fire context you
need to create an interface to work with:

```go
// wantedInterface is the interface we want our conditions & actions can work
// with.
type wantedInterface interface {
	isContextSatisfiesSomeCondition() bool
	updateSomeDataForExample(v string)
}
```

And then we create generic conditions & actions those works with it like this:

```go
type desiredSpec[T wantedInterface] struct {
	gospec.Spec[T]
}

func newDesiredSpec[T wantedInterface]() *desiredSpec[T] {
	s := &desiredSpec[T]{}
	s.Spec = gospec.New[T](s)
	return s
}

func (s *desiredSpec[T]) IsSatisfiedBy(ctx context.Context, fireContext T) (bool, error) {
	return fireContext.isContextSatisfiesSomeCondition(), nil
}

//

type desiredCondition[T wantedInterface] struct{}

func newDesiredCondition[T wantedInterface]() *desiredCondition[T] {
	return new(desiredCondition[T])
}

func (a *desiredCondition[T]) IsSatisfiedBy(ctx context.Context, fireContext T) (bool, error) {
	return fireContext.isContextSatisfiesSomeCondition(), nil
}

//

type desiredAction[T wantedInterface] struct{}

func newDesiredAction[T wantedInterface]() *desiredAction[T] {
	return new(desiredAction[T])
}

func (a *desiredAction[T]) Execute(ctx context.Context, fireContext T) error {
	fireContext.updateSomeDataForExample("some value here")
	return nil
}
```

And later we can use it with different types of fire context:

```go
type yourFireContext struct{}

func (y *yourFireContext) isContextSatisfiesSomeCondition() bool {
	return true
}

func (y *yourFireContext) updateSomeDataForExample(v string) {}

type yourAnotherFireContext struct{}

func (y *yourAnotherFireContext) isContextSatisfiesSomeCondition() bool {
	return true
}

func (y *yourAnotherFireContext) updateSomeDataForExample(v string) {}

func main() {
	r1 := krools.NewRule[*yourFireContext](
		"rule one",
		newDesiredCondition[*yourFireContext](),
		newDesiredAction[*yourFireContext](),
	)

	r2 := krools.NewRule[*yourAnotherFireContext](
		"rule one",
		newDesiredCondition[*yourAnotherFireContext](),
		newDesiredAction[*yourAnotherFireContext](),
	)

	r3 := krools.NewRule[*yourFireContext](
		"rule one",
		newDesiredSpec[*yourFireContext]().Not(),
		newDesiredAction[*yourFireContext](),
	)

	r4 := krools.NewRule[*yourAnotherFireContext](
		"rule one",
		newDesiredSpec[*yourAnotherFireContext]().Not(),
		newDesiredAction[*yourAnotherFireContext](),
	)

	_ = krools.NewKnowledgeBase[*yourFireContext]("some name").
		AddUnit("universal rules", r1, r3)

	_ = krools.NewKnowledgeBase[*yourAnotherFireContext]("some name").
		AddUnit("universal rules", r2, r4)
}
```

Or you can create a rule constructor like this:

```go
func newGenericRule[T wantedInterface](name string) *krools.Rule[T] {
	return krools.NewRule[T](name, newDesiredSpec[T]().Not(), newDesiredAction[T]())
}
```

And later reuse it lke this:

```go
func main() {
	r1 := newGenericRule[*yourFireContext]("rule one")
	r2 := newGenericRule[*yourAnotherFireContext]("rule two")

	_ = krools.NewKnowledgeBase[*yourFireContext]("some name").Add(r1)
	_ = krools.NewKnowledgeBase[*yourAnotherFireContext]("some name").Add(r2)
}
```

And now you can reuse your conditions, specs, and actions in any of fire
contexts that implements wanted interface.

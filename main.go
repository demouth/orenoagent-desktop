package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"slices"

	"github.com/demouth/orenoagent-go"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	_ "github.com/guigui-gui/guigui/basicwidget/cjkfont"
)

type modelKey int

const (
	modelKeyModel modelKey = iota
)

type Root struct {
	guigui.DefaultWidget

	background              basicwidget.Background
	createButton            basicwidget.Button
	textInput               basicwidget.TextInput
	interactionPanel        basicwidget.Panel
	interactionPanelContent interactionPanelContent

	model Model
}

func (r *Root) Model(key any) any {
	switch key {
	case modelKeyModel:
		return &r.model
	default:
		return nil
	}
}

func (r *Root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	context.SetAppScale(1.3)

	adder.AddChild(&r.background)
	adder.AddChild(&r.textInput)
	// adder.AddChild(&r.createButton)
	adder.AddChild(&r.interactionPanel)

	r.textInput.SetMultiline(true)
	r.textInput.SetAutoWrap(true)

	if r.textInput.Value() == "\n" {
		r.textInput.ForceSetValue("")
	}
	r.textInput.SetOnKeyJustPressed(func(context *guigui.Context, key ebiten.Key) {
		if key == ebiten.KeyEnter {
			if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
				ebiten.IsKeyPressed(ebiten.KeyShiftRight) ||
				ebiten.IsKeyPressed(ebiten.KeyShift) {
				// do nothing, handled below
			} else {
				r.tryAsk(r.textInput.Value())
			}
		}
	})

	// r.createButton.SetText("⇧")
	// r.createButton.SetOnUp(func(context *guigui.Context) {
	// 	r.tryAsk(r.textInput.Value())
	// })
	context.SetEnabled(&r.createButton, r.model.CanAsk(r.textInput.Value()))

	r.interactionPanelContent.SetOnAdded(func(context *guigui.Context) {
		r.interactionPanel.SetScrollOffset(0, math.Inf(-1))
	})
	r.interactionPanel.SetAutoBorder(true)
	r.interactionPanel.SetContent(&r.interactionPanelContent)
	r.interactionPanel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)

	return nil
}

func (r *Root) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&r.background, widgetBounds.Bounds())

	u := basicwidget.UnitSize(context)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items: []guigui.LinearLayoutItem{
			{
				Widget: &r.interactionPanel,
				Size:   guigui.FlexibleSize(1),
			},
			{
				Size: guigui.FixedSize(3 * u),
				Layout: guigui.LinearLayout{
					Direction: guigui.LayoutDirectionHorizontal,
					Items: []guigui.LinearLayoutItem{
						{
							Widget: &r.textInput,
							Size:   guigui.FlexibleSize(1),
						},
						// {
						// 	Widget: &r.createButton,
						// 	Size:   guigui.FixedSize(3 * u),
						// },
					},
					Gap: u / 4,
				},
			},
		},
		Gap: u,
		Padding: guigui.Padding{
			Start:  u,
			Top:    u,
			End:    u,
			Bottom: u,
		},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (r *Root) tryAsk(prompt string) {
	if r.model.TryAsk(prompt) {
		r.textInput.ForceSetValue("")
		r.interactionPanel.SetScrollOffset(0, math.Inf(-1))
	}
}

type interactionPanelContent struct {
	guigui.DefaultWidget

	placeholder basicwidget.Text

	interactionWidgets []InteractionWidget
}

const (
	interactionPanelContentEventAdded = "interactionPanelContentEventAdded"
)

func (t *interactionPanelContent) SetOnAdded(f func(context *guigui.Context)) {
	guigui.SetEventHandler(t, interactionPanelContentEventAdded, f)
}

func (t *interactionPanelContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	model := context.Model(t, modelKeyModel).(*Model)

	if model.InteractionCount() == 0 {
		t.placeholder.SetValue("Where should we begin? Ask me anything!")
		t.placeholder.SetScale(1.8)
		t.placeholder.SetMultiline(true)
		t.placeholder.SetAutoWrap(true)
		t.placeholder.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
		t.placeholder.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
		adder.AddChild(&t.placeholder)
		return nil
	}

	if model.InteractionCount() > len(t.interactionWidgets) {
		for i := len(t.interactionWidgets); i < model.InteractionCount(); i++ {
			switch model.InteractionByIndex(i).(type) {
			case *askInteraction:
				t.interactionWidgets = append(t.interactionWidgets, &promptInteractionWidget{})
			case *orenoagent.MessageDeltaResult:
				t.interactionWidgets = append(t.interactionWidgets, &outputInteractionWidget{})
			case *orenoagent.ReasoningDeltaResult:
				t.interactionWidgets = append(t.interactionWidgets, &reasoningInteractionWidget{})
			case *orenoagent.FunctionCallResult:
				t.interactionWidgets = append(t.interactionWidgets, &tooluseInteractionWidget{})
			}
		}
		guigui.DispatchEvent(t, interactionPanelContentEventAdded)

	} else {
		t.interactionWidgets = slices.Delete(t.interactionWidgets, model.InteractionCount(), len(t.interactionWidgets))
	}
	for i := range t.interactionWidgets {
		adder.AddChild(t.interactionWidgets[i])
	}

	for i := range model.InteractionCount() {
		var s string
		switch r := model.InteractionByIndex(i).(type) {
		case *askInteraction:
			s = r.String()
		case *orenoagent.MessageDeltaResult:
			s = r.String()
		case *orenoagent.ReasoningDeltaResult:
			s = r.String()
		case *orenoagent.FunctionCallResult:
			s = r.String()
		}
		if t.interactionWidgets[i].GetText() == s {
			continue
		}
		t.interactionWidgets[i].SetText(s)
		guigui.DispatchEvent(t, interactionPanelContentEventAdded)
	}
	return nil
}

func (t *interactionPanelContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	model := context.Model(t, modelKeyModel).(*Model)
	if model.InteractionCount() == 0 {
		layouter.LayoutWidget(&t.placeholder, widgetBounds.Bounds())
		// layout := guigui.LinearLayout{
		// 	Direction: guigui.LayoutDirectionHorizontal,
		// 	Items: []guigui.LinearLayoutItem{{
		// 		Widget: &t.placeholder,
		// 		Size:   guigui.FlexibleSize(1),
		// 	}},
		// }
		// layout.LayoutWidgets(context, widgetBounds.Bounds(), layouter)
		return
	}

	layout := guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Gap:       t.Gap(context),
	}
	layout.Items = make([]guigui.LinearLayoutItem, len(t.interactionWidgets))
	for i := range t.interactionWidgets {
		w := widgetBounds.Bounds().Dx()
		h := t.interactionWidgets[i].Measure(context, guigui.FixedWidthConstraints(w)).Y
		layout.Items[i] = guigui.LinearLayoutItem{
			Widget: t.interactionWidgets[i],
			Size:   guigui.FixedSize(h),
		}
	}
	layout.LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}
func (t *interactionPanelContent) Gap(context *guigui.Context) int {
	u := basicwidget.UnitSize(context)
	return int(u)
}

func (t *interactionPanelContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	model := context.Model(t, modelKeyModel).(*Model)
	if model.InteractionCount() == 0 {
		return t.placeholder.Measure(context, constraints)
	}

	var h int
	for i := range t.interactionWidgets {
		h += t.interactionWidgets[i].Measure(context, constraints).Y
		h += t.Gap(context)
	}
	w := t.DefaultWidget.Measure(context, constraints).X
	return image.Pt(w, h)
}

func main() {
	op := &guigui.RunOptions{
		Title:         "Ore-no-Agent Desktop",
		WindowMinSize: image.Pt(320, 240),
		WindowSize:    image.Pt(780, 860),
		RunGameOptions: &ebiten.RunGameOptions{
			ApplePressAndHoldEnabled: true,
		},
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

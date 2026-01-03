package main

import (
	"image"
	"image/color"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
)

type InteractionWidget interface {
	guigui.Widget
	SetText(string)
	GetText() string
}

type promptInteractionWidget struct {
	guigui.DefaultWidget

	form basicwidget.Form
	text basicwidget.Text
}

func (p *promptInteractionWidget) SetText(text string) {
	p.text.SetValue(text)
}
func (p *promptInteractionWidget) GetText() string {
	return p.text.Value()
}

func (p *promptInteractionWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddChild(&p.form)

	p.form.SetItems([]basicwidget.FormItem{
		{
			PrimaryWidget: &p.text,
		},
	})

	p.text.SetSelectable(true)
	p.text.SetMultiline(true)
	p.text.SetAutoWrap(true)
	p.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	return nil
}

func (p *promptInteractionWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items: []guigui.LinearLayoutItem{
			{
				// Widget: nil,
				Size: guigui.FlexibleSize(1),
			},
			{
				Widget: &p.form,
				Size:   guigui.FlexibleSize(3),
			},
		},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *promptInteractionWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	w, _ := constraints.FixedWidth()
	w = w * 3 / 4
	w -= u/2 + u/2 // paddings
	return image.Pt(
		w,
		p.text.Measure(context, guigui.FixedWidthConstraints(w)).Y+
			u/3+u/3, // paddings
	)
}

// outputInteractionWidget

type outputInteractionWidget struct {
	guigui.DefaultWidget

	text basicwidget.Text
}

func (t *outputInteractionWidget) SetText(text string) {
	t.text.SetValue(text)
}
func (p *outputInteractionWidget) GetText() string {
	return p.text.Value()
}

func (t *outputInteractionWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddChild(&t.text)

	t.text.SetMultiline(true)
	t.text.SetAutoWrap(true)
	t.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	return nil
}

func (t *outputInteractionWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items: []guigui.LinearLayoutItem{
			{
				Layout: guigui.LinearLayout{
					Items: []guigui.LinearLayoutItem{{
						Widget: &t.text,
						Size:   guigui.FlexibleSize(1),
					}},
					Padding: guigui.Padding{
						Top:    u / 3,
						Bottom: u / 3,
						Start:  u / 3,
						End:    u / 3,
					},
				},
			},
		},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (t *outputInteractionWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	w, _ := constraints.FixedWidth()
	w -= u/3 + u/3 // paddings
	return image.Pt(
		t.DefaultWidget.Measure(context, constraints).X,
		t.text.Measure(context, guigui.FixedWidthConstraints(w)).Y+
			u/3+u/3, // paddings
	)
}

// reasoningInteractionWidget

type reasoningInteractionWidget struct {
	guigui.DefaultWidget

	label basicwidget.Text
	text  basicwidget.Text
}

func (t *reasoningInteractionWidget) SetText(text string) {
	t.label.SetValue("Reasoning:")
	t.text.SetValue(text)
}
func (p *reasoningInteractionWidget) GetText() string {
	return p.text.Value()
}

func (t *reasoningInteractionWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddChild(&t.label)
	adder.AddChild(&t.text)

	t.label.SetScale(0.8)
	t.label.SetColor(color.RGBA{125, 125, 125, 255})
	t.label.SetBold(true)

	t.text.SetScale(0.8)
	t.text.SetColor(color.RGBA{125, 125, 125, 255})
	t.text.SetMultiline(true)
	t.text.SetAutoWrap(true)
	t.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	return nil
}

func (t *reasoningInteractionWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	w := widgetBounds.Bounds().Dx()
	h := t.label.Measure(context, guigui.FixedWidthConstraints(w)).Y
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items: []guigui.LinearLayoutItem{
			{
				Widget: &t.label,
				Size:   guigui.FixedSize(h),
			},
			{
				Layout: guigui.LinearLayout{
					Items: []guigui.LinearLayoutItem{{
						Widget: &t.text,
						Size:   guigui.FlexibleSize(1),
					}},
					Padding: guigui.Padding{
						Top:    u / 3,
						Bottom: u / 3,
						Start:  u / 3,
						End:    u / 3,
					},
				},
			},
		},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (t *reasoningInteractionWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	w, _ := constraints.FixedWidth()
	w -= u/3 + u/3 // paddings
	return image.Pt(
		t.DefaultWidget.Measure(context, constraints).X,
		t.label.Measure(context, constraints).Y+
			t.text.Measure(context, guigui.FixedWidthConstraints(w)).Y+
			u/3+u/3, // paddings
	)
}

// tooluseInteractionWidget

type tooluseInteractionWidget struct {
	reasoningInteractionWidget
}

func (t *tooluseInteractionWidget) SetText(text string) {
	t.label.SetValue("Tool Use:")
	t.text.SetValue(text)
}

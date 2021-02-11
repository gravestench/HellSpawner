package hswidget

import (
	"fmt"
	"strconv"

	"github.com/ianling/giu"
	"github.com/ianling/imgui-go"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2cof"

	"github.com/OpenDiablo2/HellSpawner/hscommon/hsenum"
)

const (
	indicatorSize = 64
)

const (
	upDownArrowW, upDownArrowH       = 15, 15
	leftRightArrowW, leftRightArrowH = 15, 15
	actionButtonW, actionButtonH     = 200, 30
	speedInputW                      = 40
)

const (
	maxSpeed = 100
)

// COFViewerState represents cof viewer's state
type COFViewerState struct {
	layerIndex     int32
	directionIndex int32
	frameIndex     int32
	state          hsenum.COFEditorState
	layer          *d2cof.CofLayer
	confirmDialog  *PopUpConfirmDialog
}

// Dispose clears viewer's layers
func (s *COFViewerState) Dispose() {
	s.layer = nil
}

// COFViewerWidget represents cof viewer's widget
type COFViewerWidget struct {
	id     string
	editor *COFEditor
	cof    *d2cof.COF
}

// COFViewer creates a cof viewer widget
func COFViewer(id string, cof *d2cof.COF, editor *COFEditor) *COFViewerWidget {
	result := &COFViewerWidget{
		id:     id,
		cof:    cof,
		editor: editor,
	}

	result.editor.cof = result.cof

	return result
}

// Build builds a cof viewer
func (p *COFViewerWidget) Build() {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	s := giu.Context.GetState(stateID)

	if s == nil {
		giu.Context.SetState(stateID, &COFViewerState{
			layer:         &p.cof.CofLayers[0],
			state:         hsenum.COFEditorStateViewer,
			confirmDialog: &PopUpConfirmDialog{},
		})

		return
	}

	state := s.(*COFViewerState)

	switch state.state {
	case hsenum.COFEditorStateViewer:
		p.buildViewer(state)
	case hsenum.COFEditorStateAddLayer:
		p.editor.makeAddLayerLayout(state).Build()
	case hsenum.COFEditorStateConfirm:
		giu.Layout{
			giu.Label("Please confirm your decision"),
			state.confirmDialog,
		}.Build()
	}
}

// nolint:funlen // no need to reduce
func (p *COFViewerWidget) buildViewer(state *COFViewerState) {
	var l1, l2, l3, l4 string

	numDirs := p.cof.NumberOfDirections
	numFrames := p.cof.FramesPerDirection

	l1 = fmt.Sprintf("Directions: %v", numDirs)

	if numDirs > 1 {
		l2 = fmt.Sprintf("Frames (x%v):", numDirs)
	} else {
		l2 = "Frames:"
	}

	// nolint:gomnd // constant
	fps := 25 * (float64(p.cof.Speed) / float64(256))
	if fps == 0 {
		fps = 25
	}

	l3 = fmt.Sprintf("FPS: %.1f", fps)
	// nolint:gomnd // miliseconds
	l4 = fmt.Sprintf("Duration: %.2fms", float64(numFrames)*(1/fps)*1000)

	layerStrings := make([]string, 0)
	for idx := range p.cof.CofLayers {
		layerStrings = append(layerStrings, strconv.Itoa(int(p.cof.CofLayers[idx].Type)))
	}

	layerList := giu.Combo("##"+p.id+"layer", layerStrings[state.layerIndex], layerStrings, &state.layerIndex).
		Size(indicatorSize).OnChange(p.onUpdate)

	directionStrings := make([]string, 0)
	for idx := range p.cof.Priority {
		directionStrings = append(directionStrings, fmt.Sprintf("%d", idx))
	}

	directionList := giu.Combo("##"+p.id+"dir", directionStrings[state.directionIndex], directionStrings, &state.directionIndex).
		Size(indicatorSize).OnChange(p.onUpdate)

	frameStrings := make([]string, 0)
	for idx := range p.cof.Priority[state.directionIndex] {
		frameStrings = append(frameStrings, fmt.Sprintf("%d", idx))
	}

	frameList := giu.Combo("##"+p.id+"frame", frameStrings[state.frameIndex], frameStrings, &state.frameIndex).
		Size(indicatorSize).OnChange(p.onUpdate)

	const vspace = 4 //nolint:unused // will be used

	speed := int32(p.cof.Speed)
	giu.TabBar("COFViewerTabs").Layout(giu.Layout{
		giu.TabItem("Animation").Layout(giu.Layout{
			giu.Label(l1),
			giu.Line(
				giu.Label(l2),
				giu.ImageButton(p.editor.leftArrowTexture).Size(leftRightArrowW, leftRightArrowH).OnClick(func() {
					if p.cof.FramesPerDirection > 0 {
						p.cof.FramesPerDirection--
					}
				}),
				giu.Label(strconv.Itoa(numFrames)),
				giu.Custom(func() {
					imgui.PopID()
					imgui.PushID("##" + p.id + "IncreaseFramesPerDirection")
				}),
				giu.ImageButton(p.editor.rightArrowTexture).Size(leftRightArrowW, leftRightArrowH).OnClick(func() {
					p.cof.FramesPerDirection++
				}),
				giu.Custom(func() {
					imgui.PopID()
					imgui.PushID("##" + p.id + "DecreaseFramesPerDirection")
				}),
			),
			giu.Line(
				giu.Label("Speed: "),
				giu.InputInt("##"+p.id+"CovViewerSpeedValue", &speed).Size(speedInputW).OnChange(func() {
					if speed <= maxSpeed {
						p.cof.Speed = int(speed)
					} else {
						p.cof.Speed = maxSpeed
					}
				}),
			),
			giu.Label(l3),
			giu.Label(l4),
		}),
		giu.TabItem("Layer").Layout(giu.Layout{
			giu.Layout{
				giu.Line(giu.Label("Selected Layer: "), layerList),
				giu.Separator(),
				p.makeLayerLayout(),
				giu.Button("Add a new layer...##"+p.id+"AddLayer").Size(actionButtonW, actionButtonH).OnClick(func() {
					state.state = hsenum.COFEditorStateAddLayer
				}),
				giu.Button("Delete current layer...##"+p.id+"DeleteLayer").Size(actionButtonW, actionButtonH).OnClick(func() {
					state.confirmDialog = NewPopUpConfirmDialog(
						"##"+p.id+"DeleteLayerConfirm",
						"Do you raly want to remove this layer?",
						"If you'll click YES, all data from this layer will be lost. Continue?",
						func() {
							p.editor.deleteCurrentLayer(state.layerIndex)
							state.state = hsenum.COFEditorStateViewer
						},
						func() {
							state.state = hsenum.COFEditorStateViewer
						},
					)

					state.state = hsenum.COFEditorStateConfirm
				}),
			},
		}),
		giu.TabItem("Priority").Layout(giu.Layout{
			giu.Line(
				giu.Label("Direction: "), directionList,
				giu.Label("Frame: "), frameList,
			),
			giu.Separator(),
			p.makeDirectionLayout(),
			giu.Button("Duplicate current direction...##"+p.id+"DuplicateDirection").Size(actionButtonW, actionButtonH).OnClick(func() {
				p.editor.duplicateDirection(state)
			}),
			giu.Button("Delete current direction...##"+p.id+"DeleteDirection").Size(actionButtonW, actionButtonH).OnClick(func() {
				NewPopUpConfirmDialog("##"+p.id+"DeleteLayerConfirm",
					"Do you raly want to remove this direction?",
					"If you'll click YES, all data from this direction will be lost. Continue?",
					func() {
						p.editor.deleteCurrentDirection(state.directionIndex)
						state.state = hsenum.COFEditorStateViewer
					},
					func() {
						state.state = hsenum.COFEditorStateViewer
					},
				)

				state.state = hsenum.COFEditorStateConfirm
			}),
		}),
	}).Build()
}

func (p *COFViewerWidget) onUpdate() {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	state := giu.Context.GetState(stateID).(*COFViewerState)

	clone := p.cof.CofLayers[state.layerIndex]
	state.layer = &clone

	giu.Context.SetState(p.id, state)
}

func (p *COFViewerWidget) makeLayerLayout() giu.Layout {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	state := giu.Context.GetState(stateID).(*COFViewerState)

	if state.layer == nil {
		p.onUpdate()
	}

	layerName := getLayerName(state.layer.Type)

	strType := fmt.Sprintf("Type: %s (%s)", state.layer.Type, layerName)
	strShadow := fmt.Sprintf("Shadow: %t", state.layer.Shadow > 0)
	strSelectable := fmt.Sprintf("Selectable: %t", state.layer.Selectable)
	strTransparent := fmt.Sprintf("Transparent: %t", state.layer.Transparent)

	effect := hsenum.GetDrawEffectName(state.layer.DrawEffect)

	strEffect := fmt.Sprintf("Draw Effect: %s", effect)

	weapon := hsenum.GetWeaponClassString(state.layer.WeaponClass)

	strWeaponClass := fmt.Sprintf("Weapon Class: (%s) %s", state.layer.WeaponClass, weapon)

	return giu.Layout{
		giu.Label(strType),
		giu.Label(strShadow),
		giu.Label(strSelectable),
		giu.Label(strTransparent),
		giu.Label(strEffect),
		giu.Label(strWeaponClass),
	}
}

// nolint:gocyclo // can't reduce
func getLayerName(i interface{}) string {
	var t d2enum.CompositeType

	switch j := i.(type) {
	case int:
		t = d2enum.CompositeType(j)
	case d2enum.CompositeType:
		t = j
	}

	var layerName string

	switch t {
	case d2enum.CompositeTypeHead:
		layerName = "Head"
	case d2enum.CompositeTypeTorso:
		layerName = "Torso"
	case d2enum.CompositeTypeLegs:
		layerName = "Legs"
	case d2enum.CompositeTypeRightArm:
		layerName = "Right Arm"
	case d2enum.CompositeTypeLeftArm:
		layerName = "Left Arm"
	case d2enum.CompositeTypeRightHand:
		layerName = "Right Hand"
	case d2enum.CompositeTypeLeftHand:
		layerName = "Left Hand"
	case d2enum.CompositeTypeShield:
		layerName = "Shield"
	case d2enum.CompositeTypeSpecial1:
		layerName = "Special 1"
	case d2enum.CompositeTypeSpecial2:
		layerName = "Special 2"
	case d2enum.CompositeTypeSpecial3:
		layerName = "Special 3"
	case d2enum.CompositeTypeSpecial4:
		layerName = "Special 4"
	case d2enum.CompositeTypeSpecial5:
		layerName = "Special 5"
	case d2enum.CompositeTypeSpecial6:
		layerName = "Special 6"
	case d2enum.CompositeTypeSpecial7:
		layerName = "Special 7"
	case d2enum.CompositeTypeSpecial8:
		layerName = "Special 8"
	}

	return layerName
}

func (p *COFViewerWidget) makeDirectionLayout() giu.Layout {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	state := giu.Context.GetState(stateID).(*COFViewerState)

	frames := p.cof.Priority[state.directionIndex]
	layers := frames[int(state.frameIndex)%len(frames)]

	return giu.Layout{
		giu.Label("Render Order (first to last):"),
		giu.Custom(func() {
			for idx := range layers {
				currentIdx := idx
				giu.Line(
					giu.ImageButton(p.editor.upArrowTexture).Size(upDownArrowW, upDownArrowH).OnClick(func() {
						if currentIdx > 0 {
							p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx-1],
								p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx] =
								p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx],
								p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx-1]
						}
					}),
					giu.Custom(func() {
						imgui.PopID()
						imgui.PushID(fmt.Sprintf("LayerPriorityUp_%d", currentIdx))
					}),
					giu.ImageButton(p.editor.downArrowTexture).Size(upDownArrowW, upDownArrowH).OnClick(func() {
						if currentIdx < len(layers)-1 {
							p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx],
								p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx+1] =
								p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx+1],
								p.cof.Priority[state.directionIndex][state.frameIndex][currentIdx]
						}
					}),
					giu.Custom(func() {
						imgui.PopID()
						imgui.PushID(fmt.Sprintf("LayerPriorityDown_%d", currentIdx))
					}),
					giu.Label(fmt.Sprintf("%d: %s", idx, getLayerName(layers[idx]))),
				).Build()
			}
		}),
	}
}

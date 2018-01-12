package peterjsons

import (
	"errors"
	"fmt"

	"github.com/Jeffail/gabs"
)

var (

	//ErrParseMaterial is the error given when fails to parse material data
	ErrParseMaterial = errors.New("Failed to parse material data")
	//ErrParseRecipe is the error given when fails to parse recipe data
	ErrParseRecipe = errors.New("Failed to parse recipe data")
	//ErrParseRecipeStructure is the error given when fails to parse recipe structure data
	ErrParseRecipeStructure = errors.New("Failed to parse recipe structure")
	//ErrRecipeStudy is the error given when fails to study recipe structure data
	ErrRecipeStudy = errors.New("Failed to study recipe structure")

	//RecipeKey Reserved name for recipe object
	RecipeKey = "recipe"
	//ActionSwap Reserved name for swap actions
	ActionSwap = "swap"
	//ActionConcat Reserved name for concat actions
	ActionConcat = "concat"
	//ActionSwapMap Reserved name for swap map actions
	ActionSwapMap = "swapmap"

	//SubActionValue Reserved name for values sub actions
	SubActionValue = "value"
	//SubActionReplace Reserved name for replace sub actions
	SubActionReplace = "replace"
	//SubActionFormat Reserved name for format sub actions
	SubActionFormat = "format"
)

//Peterjsons is a structure that holds the required material and recipe
//granted with cook and output in multiple forms behaviors
type Peterjsons struct {
	material *gabs.Container
	recipe   *gabs.Container
	result   *gabs.Container
}

//New returns a new Peterjsons instance
//Receives a material and recipe []bytes containing material(input) and recipe json structures
func New(material, recipe []byte) (*Peterjsons, error) {
	m, err := gabs.ParseJSON(material)
	if err != nil {
		return nil, ErrParseMaterial
	}

	r, err := gabs.ParseJSON(recipe)
	if err != nil {
		return nil, ErrParseRecipe
	}

	return &Peterjsons{m, r, gabs.New()}, err
}

//Cook applys the studies done to the material
//and the rules described on the recipe to a output a form of result
func (peter *Peterjsons) Cook() error {
	studies, err := peter.study()
	if err != nil {
		return ErrRecipeStudy
	}

	return peter.apply(studies...)
}

//JSONResult outputs the cook result in JSON format
func (peter *Peterjsons) JSONResult() string {
	return string(peter.result.EncodeJSON())
}

//BytesResult outputs the cook result in Bytes format
func (peter *Peterjsons) BytesResult() []byte {
	return peter.result.Bytes()
}

// --------------------------------------------------------------------------------------------------------------------------------//

func (peter *Peterjsons) study() ([]action, error) {
	recipeData, err := peter.recipe.S(RecipeKey).ChildrenMap()
	actions := make([]action, 0)

	if err != nil {
		return actions, ErrParseRecipeStructure
	}

	for key, object := range recipeData {
		switch object.Data().(type) {
		case int:
		case string:
			actions = append(actions, studyInterfaceRecipe(key, object.Data(), peter.material))
		case map[string]interface{}:
			actions = append(actions, studyMapInterfaceRecipe(key, object.Data().(map[string]interface{}), peter.recipe)...)
		}
	}

	return actions, nil
}

func (peter *Peterjsons) apply(actions ...action) error {
	for index := range actions {
		switch actions[index].Operation {
		case ActionSwap:
			swapAction(actions[index], peter.material, peter.result)
		case ActionConcat:
			concatAction(actions[index], peter.material, peter.result)
		case ActionSwapMap:
			swapmapAction(actions[index], peter.material, peter.result)
		}
	}

	return nil
}

//------------------------------------------------------STUDIES-------------------------------------------------------------------//

func studyInterfaceRecipe(key string, object interface{}, material *gabs.Container) action {
	replaces := make([]interface{}, 0)
	return action{
		Context:   key,
		Operation: "swap",
		Replaces:  append(replaces, object),
	}
}

func studyMapInterfaceRecipe(key string, object map[string]interface{}, material *gabs.Container) []action {
	actions := make([]action, 0)
	auxactions := make(map[string]*action, 0)
	for objectIndex := range object {
		switch objectIndex {
		case "operation", "*", "map", "values", "value", "replace", "replaces", "separator", "format":
			auxactions = studyMapInterfaceAction(key, objectIndex, auxactions, material)
		default:
			actions = studyMapInterfaceActions(key, objectIndex, actions, material)
		}
	}

	for aux := range auxactions {
		actions = append(actions, *auxactions[aux])
	}

	return actions
}

func studyMapInterfaceAction(key, objectIndex string, auxactions map[string]*action, material *gabs.Container) map[string]*action {
	objectMap := material.Path(fmt.Sprintf("%s.%s", RecipeKey, key)).Data()
	if auxactions[key] == nil {
		auxactions[key] = &action{}
	}

	if objectIndex == "values" {
		auxactions[key].setParameter(objectIndex, objectMap.(map[string]interface{})[objectIndex].([]interface{})[0])
		auxactions[key].setParameter("context", key)
		return auxactions
	}

	auxactions[key].setParameter(objectIndex, objectMap.(map[string]interface{})[objectIndex])
	auxactions[key].setParameter("context", key)

	return auxactions
}

func studyMapInterfaceActions(key, objectIndex string, actions []action, material *gabs.Container) []action {
	objectMap := material.Path(fmt.Sprintf("%s.%s.%s", RecipeKey, key, objectIndex)).Data()
	studyAction := &action{}
	switch mapType := objectMap.(type) {
	case map[string]interface{}:
		for mapKey := range mapType {
			switch objectType := mapType[mapKey].(type) {
			case []interface{}:
				for value := range objectType {
					studyAction.setParameter(mapKey, objectType[value])
				}
			default:
				studyAction.setParameter(mapKey, objectType)
			}

			studyAction.setParameter("context", fmt.Sprintf("%s.%s", key, objectIndex))
		}
	}

	return append(actions, *studyAction)
}

//---------------------------------------------------Actions----------------------------------------------------------------------//

func concatAction(action action, material, result *gabs.Container) {
	separatedString := ""
	for value := range action.Values {
		gatheredvalue := material.Path(action.Values[value].(string)).Data()
		separatedString = fmt.Sprintf("%s%v", separatedString, gatheredvalue)
		if value < len(action.Values)-1 {
			separatedString = fmt.Sprintf("%s%s", separatedString, action.Separator)
		}
	}

	result.SetP(separatedString, action.Context)
}

func swapAction(action action, material, result *gabs.Container) {
	if len(action.Values) == 0 {
		swapSingleAction(action, material, result)
		return
	}

	swapMultipleActions(action, material, result)
}

func swapmapAction(action action, material, result *gabs.Container) {
	objects := make([]interface{}, 0)
	for swaps := range action.Values {
		materialValues := material.Path(action.Values[swaps].(string)).Data().([]interface{})

		for materialIndex := range materialValues {
			object := make(map[string]interface{}, 0)
			for mapIndex := range action.Map {
				for mapValueIndex := range action.Map[mapIndex].(map[string]interface{}) {
					if action.Map[mapIndex].(map[string]interface{})[mapValueIndex].(map[string]interface{})[SubActionReplace] != nil {
						object[mapValueIndex] = action.Map[mapIndex].(map[string]interface{})[mapValueIndex].(map[string]interface{})[SubActionReplace]
						continue
					}

					if action.Map[mapIndex].(map[string]interface{})[mapValueIndex].(map[string]interface{})[SubActionValue] != nil {
						object[mapValueIndex] = materialValues[materialIndex].(map[string]interface{})[action.Map[mapIndex].(map[string]interface{})[mapValueIndex].(map[string]interface{})[SubActionValue].(string)]
					}
				}
				objects = append(objects, object)
			}
		}
	}

	result.SetP(objects, action.Context)
}

func swapSingleAction(action action, material, result *gabs.Container) {
	gatheredvalue := material.Path(action.Replaces[0].(string)).Data()

	if action.Format != "" {
		result.SetP(fmt.Sprintf(action.Format, gatheredvalue), action.Context)
		return
	}

	result.SetP(gatheredvalue, action.Context)
	return
}

func swapMultipleActions(action action, material, result *gabs.Container) {
	for value := range action.Values {
		valuePath := action.Values[value].(string)
		gatheredvalue := material.Path(valuePath).Data()

		if len(action.Replaces) > 0 && action.Format != "" {
			result.SetP(fmt.Sprintf(action.Format, action.Replaces[value]), action.Context)
			continue
		}

		if action.Format != "" {
			result.SetP(fmt.Sprintf(action.Format, gatheredvalue), action.Context)
			continue
		}

		if len(action.Replaces) > 0 {
			result.SetP(action.Replaces[value], action.Context)
			continue
		}

		result.SetP(gatheredvalue, action.Context)
	}
}

//---------------------------------------------------Action Structure------------------------------------------------------------//

type action struct {
	Context   string
	Values    []interface{}
	Operation string
	Format    string
	Separator string
	Replaces  []interface{}
	Map       []interface{}
}

func (a *action) setParameter(context string, value interface{}) {
	switch context {
	case "operation":
		a.Operation = value.(string)
	case "format":
		a.Format = value.(string)
	case "separator":
		a.Separator = value.(string)
	case "context":
		a.Context = value.(string)
	case "values", "value":
		a.Values = append(a.Values, value)
	case "replaces", "replace":
		a.Replaces = append(a.Replaces, value)
	case "*", "all":
		a.Map = append(a.Map, value)
	}
}

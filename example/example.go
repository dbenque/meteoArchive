package meteoExample

import (
	"fmt"
	"meteo/infoclimat"
	"meteo/meteoAPI"
	"reflect"

	"code.google.com/p/biogo.store/kdtree"
)

func Example() {

	paris := meteoAPI.NewStation("Paris", 42, 48.8534100, 2.3488000)
	mon34 := meteoAPI.NewStation("Paris", 42, 43.6109200, 3.8772300)

	testStation := [...]*meteoAPI.Station{paris, mon34}
	//ExampleT_InRange()

	stations, _ := infoclimat.GetStations()

	tree := kdtree.New(stations, true)
	for _, s := range testStation {
		keeper := kdtree.NewDistKeeper(100)
		tree.NearestSet(keeper, s)
		fmt.Println("Found :", keeper.Len())

		for keeper.Len() > 0 {
			v := keeper.Heap.Pop()

			if c, ok := v.(kdtree.ComparableDist); !ok {
				fmt.Println("bad type (expect kdtree.ComparableDist) :", reflect.TypeOf(v))
			} else {
				if s, ok := c.Comparable.(meteoAPI.Station); !ok {
					fmt.Println("bad type (expect Station):", reflect.TypeOf(c))
				} else {
					fmt.Println("Station at ", int(c.Dist), " km: ", s.Name)
				}
			}
		}
	}
}

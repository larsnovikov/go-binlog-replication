package parser

import (
	"fmt"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"go-binlog-replication/src/helpers"
	"go-binlog-replication/src/models"
	"log"
	"runtime/debug"
	"strconv"
)

type binlogHandler struct {
	canal.DummyEventHandler
	BinlogParser
}

var curPosition mysql.Position

func canOperateTable(tableName string) bool {
	for _, v := range helpers.GetTables() {
		if v == tableName {
			return true
		}
	}

	return false
}

func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print(r, " ", string(debug.Stack()))
		}
	}()

	if curPosition.Pos == 0 {
		curPosition = getMasterPos()
	}

	var n int
	var k int

	if canOperateTable(e.Table.Name) == false {
		return nil
	}

	model := models.GetModel(e.Table.Name)

	switch e.Action {
	case canal.DeleteAction:
		for _, row := range e.Rows {
			model.ParseKey(row)
			if model.Delete() == true {
				setMasterPosFromCanal(curPosition)
				fmt.Printf("[%s] is deleted \n", e.Table)
			}
		}

		return nil
	case canal.UpdateAction:
		n = 1
		k = 2
	case canal.InsertAction:
		n = 0
		k = 1
	}

	for i := n; i < len(e.Rows); i += k {
		key := e.Table.Schema + "." + e.Table.Name
		switch key {
		case model.SchemaName() + "." + model.TableName():
			h.GetBinLogData(model, e, i)

			if e.Action == canal.UpdateAction {
				oldModel := models.GetModel(e.Table.Name)
				h.GetBinLogData(oldModel, e, i-1)
				if model.Update() == true {
					setMasterPosFromCanal(curPosition)
					fmt.Printf("[%s] update row\n", e.Table)
				}
			} else {
				if model.Insert() == true {
					setMasterPosFromCanal(curPosition)
					fmt.Printf("[%s] insert row\n", e.Table)
				}
			}
		}

	}
	return nil
}

func (h *binlogHandler) String() string {
	return "binlogHandler"
}

func BinlogListener() {
	helpers.MakeCredentials()

	c, err := getDefaultCanal()
	if err == nil {
		coords, err := getMasterPosFromCanal(c)
		if err == nil {
			c.SetEventHandler(&binlogHandler{})
			c.RunFrom(coords)
		}
	}
}

func getMasterPosFromCanal(c *canal.Canal) (mysql.Position, error) {
	// try to get coords from storage
	position, err := strconv.ParseUint(models.GetValue(helpers.LastPositionPos), 10, 32)
	if err == nil {
		pos := mysql.Position{
			models.GetValue(helpers.LastPositionName),
			uint32(position),
		}

		if pos.Pos != 0 && pos.Name != "" {
			showPos(pos)
			return pos, nil
		}
	}

	// get coords from mysql
	pos, err := c.GetMasterPos()
	showPos(pos)

	return pos, err
}

func setMasterPosFromCanal(position mysql.Position) {
	// save position
	models.SetValue(helpers.LastPositionPos, fmt.Sprint(position.Pos))
	models.SetValue(helpers.LastPositionName, position.Name)

	curPosition = position
}

func getMasterPos() mysql.Position {
	c, err := getDefaultCanal()
	if err != nil {
		log.Fatal("Invalid canal")
	}

	coords, err := getMasterPosFromCanal(c)
	if err != nil {
		log.Fatal("Invalid pos")
	}

	return coords
}

func getDefaultCanal() (*canal.Canal, error) {
	master := helpers.GetCredentials(helpers.DBMaster)

	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", master.Host, master.Port)
	cfg.User = master.User
	cfg.Password = master.Pass
	cfg.Flavor = master.Type

	cfg.Dump.ExecutionPath = ""

	return canal.NewCanal(cfg)
}

func showPos(pos mysql.Position) {
	fmt.Println(fmt.Sprintf("Get pos from storage. Pos: %s; Name: %s", fmt.Sprint(pos.Pos), pos.Name))
}
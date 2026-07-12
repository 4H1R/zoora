// Command gen-import-templates regenerates the static import template files
// served by the frontend. Run from the repo root:
//
//	go run ./scripts/gen-import-templates
package main

import (
	"log"

	"github.com/xuri/excelize/v2"
)

func main() {
	users := excelize.NewFile()
	must(users.SetSheetName(users.GetSheetList()[0], "Users"))
	must(users.SetSheetRow("Users", "A1", &[]string{"name", "username", "password", "role"}))
	must(users.SetSheetRow("Users", "A2", &[]string{"Ali Rezaei", "ali.rezaei", "", "Student"}))
	must(users.SetSheetRow("Users", "A3", &[]string{"Sara Karimi", "sara.k", "mypassword1", "-"}))
	must(users.SaveAs("frontend/public/templates/users-import-template.xlsx"))

	classes := excelize.NewFile()
	must(classes.SetSheetName(classes.GetSheetList()[0], "Classes"))
	_, err := classes.NewSheet("Members")
	must(err)
	must(classes.SetSheetRow("Classes", "A1", &[]string{"class_name", "owner_username", "description", "capacity"}))
	must(classes.SetSheetRow("Classes", "A2", &[]string{"Math-A", "ali.rezaei", "Algebra basics", "30"}))
	must(classes.SetSheetRow("Members", "A1", &[]string{"class_name", "member_username"}))
	must(classes.SetSheetRow("Members", "A2", &[]string{"Math-A", "sara.k"}))
	must(classes.SaveAs("frontend/public/templates/classes-import-template.xlsx"))

	members := excelize.NewFile()
	must(members.SetSheetName(members.GetSheetList()[0], "Members"))
	must(members.SetSheetRow("Members", "A1", &[]string{"class_name", "member_username"}))
	must(members.SetSheetRow("Members", "A2", &[]string{"Math-A", "sara.k"}))
	must(members.SaveAs("frontend/public/templates/class-members-import-template.xlsx"))

	log.Println("templates written to frontend/public/templates/")
}

func must(v any) {
	switch e := v.(type) {
	case error:
		if e != nil {
			log.Fatal(e)
		}
	}
}

package bigip

import (
	"fmt"
	"testing"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TestPpUniversalName = fmt.Sprintf("/%s/test-ppuniversal", TestPartition)

var TestPpUniversalResource = `
resource "bigip_ltm_persistence_profile_universal" "test_ppuniversal" {
	name = "` + TestPpUniversalName + `"
	defaults_from = "/Common/universal"
	rule = "/Common/irule_custom_persistence_universal"
}
`

func TestAccBigipLtmPersistenceProfileUniversalCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(testCheckBigipLtmPersistenceProfileUniversalDestroyed),
		Steps: []resource.TestStep{
			{
				Config: TestPpUniversalResource,
				Check: resource.ComposeTestCheckFunc(
					testBigipLtmPersistenceProfileUniversalExists(TestPpUniversalName, true),
					resource.TestCheckResourceAttr("bigip_ltm_persistence_profile_universal.test_ppuniversal", "name", TestPpUniversalName),
					resource.TestCheckResourceAttr("bigip_ltm_persistence_profile_universal.test_ppuniversal", "defaults_from", "/Common/universal"),
					resource.TestCheckResourceAttr("bigip_ltm_persistence_profile_universal.test_ppuniversal", "rule", "/Common/irule_custom_persistence_universal"),
				),
			},
		},
	})

}

func TestAccBigipLtmPersistenceProfileUniversalImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAcctPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckBigipLtmPersistenceProfileUniversalDestroyed,
		Steps: []resource.TestStep{
			{
				Config: TestPpUniversalResource,
				Check: resource.ComposeTestCheckFunc(
					testBigipLtmPersistenceProfileUniversalExists(TestPpUniversalName, true),
				),
				ResourceName:      TestPpUniversalName,
				ImportState:       false,
				ImportStateVerify: true,
			},
		},
	})
}

func testBigipLtmPersistenceProfileUniversalExists(name string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*bigip.BigIP)

		pp, err := client.GetUniversalPersistenceProfile(name)
		if err != nil {
			return err
		}
		if exists && pp == nil {
			return fmt.Errorf("Universal Persistence Profile %s does not exist.", name)
		}
		if !exists && pp != nil {
			return fmt.Errorf("Universal Persistence Profile %s exists.", name)
		}
		return nil
	}
}

func testCheckBigipLtmPersistenceProfileUniversalDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*bigip.BigIP)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bigip_ltm_persistence_profile_universal" {
			continue
		}

		name := rs.Primary.ID
		pp, err := client.GetSourceAddrPersistenceProfile(name)
		if err != nil {
			return err
		}

		if pp != nil {
			return fmt.Errorf("Universal Persistence Profile %s not destroyed.", name)
		}
	}
	return nil
}

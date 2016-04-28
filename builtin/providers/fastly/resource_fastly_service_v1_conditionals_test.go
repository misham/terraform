package fastly

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly"
)

func TestAccFastlyServiceV1_conditional_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("%s.notadomain.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccServiceV1ConditionConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1ConditionalAttributes(&service, name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "condition.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1ConditionalAttributes(service *gofastly.ServiceDetail, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Conditions for (%s), version (%s): %s", service.Name, service.ActiveVersion.Number, err)
		}

		log.Printf("\n@@@\nFound conditions: %#v\n@@@\n", conditionList)

		// var deleted []string
		// var added []string
		// for _, h := range headersList {
		// 	if h.Action == gofastly.HeaderActionDelete {
		// 		deleted = append(deleted, h.Destination)
		// 	}
		// 	if h.Action == gofastly.HeaderActionSet {
		// 		added = append(added, h.Destination)
		// 	}
		// }

		// sort.Strings(headersAdded)
		// sort.Strings(headersDeleted)
		// sort.Strings(deleted)
		// sort.Strings(added)

		// if !reflect.DeepEqual(headersDeleted, deleted) {
		// 	return fmt.Errorf("Deleted Headers did not match.\n\tExpected: (%#v)\n\tGot: (%#v)", headersDeleted, deleted)
		// }
		// if !reflect.DeepEqual(headersAdded, added) {
		// 	return fmt.Errorf("Added Headers did not match.\n\tExpected: (%#v)\n\tGot: (%#v)", headersAdded, added)
		// }

		return nil
	}
}

func testAccServiceV1ConditionConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  header {
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  condition {
    name = "some amz condition"
    type = "REQUEST"

    statement = <<EOF
req.url ~ "^/yolo/"
EOF

    priority = 10
  }

  force_destroy = true
}`, name, domain)
}

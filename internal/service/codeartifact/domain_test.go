package codeartifact_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodeartifact "github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
)

func init() {
	resource.AddTestSweepers("aws_codeartifact_domain", &resource.Sweeper{
		Name: "aws_codeartifact_domain",
		F:    testSweepCodeArtifactDomains,
	})
}

func testSweepCodeArtifactDomains(region string) error {
	client, err := acctest.SharedRegionalSweeperClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CodeArtifactConn
	input := &codeartifact.ListDomainsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListDomainsPages(input, func(page *codeartifact.ListDomainsOutput, lastPage bool) bool {
		for _, domainPtr := range page.Domains {
			if domainPtr == nil {
				continue
			}

			domain := aws.StringValue(domainPtr.Name)
			input := &codeartifact.DeleteDomainInput{
				Domain: domainPtr.Name,
			}

			log.Printf("[INFO] Deleting CodeArtifact Domain: %s", domain)

			_, err := conn.DeleteDomain(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeArtifact Domain (%s): %w", domain, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if acctest.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeArtifact Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeArtifact Domains: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccCodeArtifactDomain_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("domain/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttr(resourceName, "asset_size_bytes", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository_count", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_key", "aws_kms_key.test", "arn"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeArtifactDomain_defaultEncryptionKey(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("codeartifact", t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDefaultEncryptionKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("domain/%s", rName)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "encryption_key", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttr(resourceName, "asset_size_bytes", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository_count", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeArtifactDomain_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("codeartifact", t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccCodeArtifactDomain_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDefaultEncryptionKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodeartifact.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact domain set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn

		domainOwner, domainName, err := tfcodeartifact.DecodeDomainID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeDomain(&codeartifact.DescribeDomainInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
		})

		return err
	}
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_domain" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn

		domainOwner, domainName, err := tfcodeartifact.DecodeDomainID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.DescribeDomain(&codeartifact.DescribeDomainInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
		})

		if err == nil {
			if aws.StringValue(resp.Domain.Arn) == rs.Primary.ID {
				return fmt.Errorf("CodeArtifact Domain %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrMessageContains(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccDomainBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}
`, rName)
}

func testAccDomainTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codeartifact_domain" "test" {
  domain = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDomainTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codeartifact_domain" "test" {
  domain = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDomainDefaultEncryptionKeyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codeartifact_domain" "test" {
  domain = %[1]q
}
`, rName)
}

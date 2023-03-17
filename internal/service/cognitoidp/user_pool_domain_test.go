package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCognitoIDPUserPoolDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_distribution"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_distribution_arn"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_distribution_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttrSet(resourceName, "s3_bucket"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccCognitoIDPUserPoolDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPoolDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDomain_custom(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	poolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	acmCertificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_custom(rootDomain, domain, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_distribution"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_distribution_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrPair(resourceName, "domain", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttrSet(resourceName, "s3_bucket"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccCognitoIDPUserPoolDomain_customCertUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	poolName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	acmInitialValidationResourceName := "aws_acm_certificate_validation.initial_test"
	acmUpdatedValidationResourceName := "aws_acm_certificate_validation.updated_test"
	acmInitialCertResourceName := "aws_acm_certificate.initial"
	acmUpdatedCertResourceName := "aws_acm_certificate.updated"
	cognitoPoolResourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDomainConfig_customCertUpdate(rootDomain, domain, poolName, acmInitialValidationResourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainExists(ctx, cognitoPoolResourceName),
					testAccCheckUserPoolDomainCertMatches(ctx, cognitoPoolResourceName, acmInitialCertResourceName),
					resource.TestCheckResourceAttrPair(cognitoPoolResourceName, "certificate_arn", acmInitialCertResourceName, "arn"),
				),
			},
			{
				Config: testAccUserPoolDomainConfig_customCertUpdate(rootDomain, domain, poolName, acmUpdatedValidationResourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolDomainCertMatches(ctx, cognitoPoolResourceName, acmUpdatedCertResourceName),
					resource.TestCheckResourceAttrPair(cognitoPoolResourceName, "certificate_arn", acmUpdatedCertResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckUserPoolDomainExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		_, err := tfcognitoidp.FindUserPoolDomain(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckUserPoolDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_domain" {
				continue
			}

			_, err := tfcognitoidp.FindUserPoolDomain(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User Pool Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserPoolDomainCertMatches(ctx context.Context, cognitoResourceName, certResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cognitoResource, ok := s.RootModule().Resources[cognitoResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", cognitoResourceName)
		}

		if cognitoResource.Primary.ID == "" {
			return errors.New("No Cognito User Pool Domain ID is set")
		}

		certResource, ok := s.RootModule().Resources[certResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", cognitoResourceName)
		}

		if certResource.Primary.ID == "" {
			return errors.New("No ACM Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn()

		domain, err := tfcognitoidp.FindUserPoolDomain(ctx, conn, cognitoResource.Primary.ID)

		if err != nil {
			return err
		}

		if domain.CustomDomainConfig == nil {
			return fmt.Errorf("No Custom Domain set on Cognito User Pool: %s", aws.StringValue(domain.UserPoolId))
		}

		if aws.StringValue(domain.CustomDomainConfig.CertificateArn) != certResource.Primary.ID {
			return fmt.Errorf("Certificate ARN on Custom Domain does not match, expected: %s, got: %s", certResource.Primary.ID, aws.StringValue(domain.CustomDomainConfig.CertificateArn))
		}

		return nil
	}
}

func testAccUserPoolDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccUserPoolDomainConfig_custom(rootDomain string, domain string, poolName string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

resource "aws_cognito_user_pool" "test" {
  name = %[3]q
}

resource "aws_cognito_user_pool_domain" "test" {
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  domain          = aws_acm_certificate.test.domain_name
  user_pool_id    = aws_cognito_user_pool.test.id
}
`, rootDomain, domain, poolName)
}

func testAccUserPoolDomainConfig_customCertUpdate(rootDomain string, domain string, poolName string, appliedCertValidation string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "initial" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_acm_certificate" "updated" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_route53_record" "initial_test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.initial.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.initial.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.initial.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_route53_record" "updated_test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.updated.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.updated.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.updated.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "initial_test" {
  certificate_arn         = aws_acm_certificate.initial.arn
  validation_record_fqdns = [aws_route53_record.initial_test.fqdn]
}
resource "aws_acm_certificate_validation" "updated_test" {
  certificate_arn         = aws_acm_certificate.updated.arn
  validation_record_fqdns = [aws_route53_record.updated_test.fqdn]
}

resource "aws_cognito_user_pool" "test" {
  name = %[3]q
}

resource "aws_cognito_user_pool_domain" "test" {
  certificate_arn = %[4]s.certificate_arn
  domain          = aws_acm_certificate.initial.domain_name
  user_pool_id    = aws_cognito_user_pool.test.id
}
`, rootDomain, domain, poolName, appliedCertValidation)
}

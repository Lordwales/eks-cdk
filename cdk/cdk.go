package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseks"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"cdk/utils"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	Log "github.com/sirupsen/logrus"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := awsec2.NewVpc(stack, jsii.String("MaxVpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2), // This creates subnets in 2 AZs
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
		},
	})

	cluster := awseks.NewCluster(stack, jsii.String("MaxCluster"), &awseks.ClusterProps{
		Vpc:             vpc,
		DefaultCapacity: jsii.Number(0),
		Version:         awseks.KubernetesVersion_V1_25(),
		VpcSubnets: &[]*awsec2.SubnetSelection{
			{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		},
		ClusterLogging: &[]awseks.ClusterLoggingTypes{awseks.ClusterLoggingTypes_API, awseks.ClusterLoggingTypes_AUDIT,
			awseks.ClusterLoggingTypes_AUTHENTICATOR, awseks.ClusterLoggingTypes_SCHEDULER, awseks.ClusterLoggingTypes_CONTROLLER_MANAGER},
	})

	cluster.AddFargateProfile(jsii.String("MaxFargateProfile"), &awseks.FargateProfileOptions{
		Selectors: &[]*awseks.Selector{{
			Namespace: jsii.String("default"),
		}},
	})

	sg := awsec2.NewSecurityGroup(stack, jsii.String("EKSFargateSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(true),
		SecurityGroupName: jsii.String("EKSFargateSG"),
	})

	// Ingress rules for external traffic
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("Allow internal VPC traffic"), jsii.Bool(false))
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(443)), jsii.String("Allow internal VPC traffic"), jsii.Bool(false))
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_IcmpPing(), jsii.String("Allow ICMP (ping) from everywhere"), jsii.Bool(false))

	// Allow all internal TCP traffic within VPC
	sg.AddIngressRule(awsec2.Peer_Ipv4(vpc.VpcCidrBlock()), awsec2.Port_TcpRange(jsii.Number(0), jsii.Number(65535)), jsii.String("Allow all TCP traffic within VPC"), jsii.Bool(false))
	sg.AddIngressRule(awsec2.Peer_Ipv4(vpc.VpcCidrBlock()), awsec2.Port_UdpRange(jsii.Number(0), jsii.Number(65535)), jsii.String("Allow all UDP traffic within VPC"), jsii.Bool(false))

	cluster.Connections().AddSecurityGroup(sg)

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("CdkQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	return stack
}

func main() {
	defer jsii.Close()
	utils.InitLogs(nil, Log.DebugLevel)

	loggerLevel := os.Getenv("LOG_LEVEL") // *******###### Needs Looking Into ****########
	switch loggerLevel {
	case "debug":
		utils.Logger.SetLevel(Log.DebugLevel)
	case "error":
		utils.Logger.SetLevel(Log.ErrorLevel)
	case "fatal":
		utils.Logger.SetLevel(Log.FatalLevel)
	case "info":
		utils.Logger.SetLevel(Log.InfoLevel)
	case "warn":
		utils.Logger.SetLevel(Log.WarnLevel)
	default:
		utils.Logger.SetLevel(Log.DebugLevel)
	}
	// utils.Logger.WithFields(Log.Fields{"properties": p.String()}).Debug("properties")
	app := awscdk.NewApp(nil)

	NewCdkStack(app, "CdkStack", &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	// return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}

// func getEnv(key, fallback string) string {
// 	if value, ok := os.LookupEnv(key); ok {
// 		return value
// 	}
// 	return fallback
// }

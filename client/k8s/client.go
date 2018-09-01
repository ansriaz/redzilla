package k8s

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	// "k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	// "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/dynamic"
	// "k8s.io/apimachinery/pkg/util/yaml"
	// "k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/discovery"
	// "k8s.io/client-go/dynamic"
	// "k8s.io/apimachinery/pkg/util/yaml"
	// "k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// "k8s.io/apimachinery/pkg/api/meta"
	// "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ansriaz/redzilla/client"
	"github.com/ansriaz/redzilla/model"
)

var decode = scheme.Codecs.UniversalDeserializer().Decode

type K8SClient struct {
	client.Client
	config     *model.Config
	k8sconfig  *rest.Config
	client     *kubernetes.Clientset
	k8sclient  dynamic.Interface
	restMapper meta.RESTMapper
}

func NewK8SClient() *K8SClient {
	return &K8SClient{}
}

func (c *K8SClient) Init(conf *model.Config) error {
	// TODO settings for outside of the cluster
	c.config = conf

	var config *rest.Config
	var err error
	if strings.EqualFold(conf.ClusterAccess, "Incluster") {
		log.Info("Using inCluster access kind")
		config, err = rest.InClusterConfig() // Gets the configuration from the running pod
		if err != nil {
			return err
		}
	} else {
		log.Infof("Using kubeconfig \"%s\" access kind", conf.ClusterAccess)

		// cwd, err := os.Getwd()
		// if err != nil {
		// 	return err
		// }

		// config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(cwd, conf.ClusterAccess))
		config, err = clientcmd.BuildConfigFromFlags("", "/home/gilardoni/.kube/config")
		config, err = clientcmd.BuildConfigFromFlags("", conf.ClusterAccess)
		if err != nil {
			return err
		}
	}

	client1, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	client2, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.k8sclient = client1
	c.client = client2

	// dd := c.k8sclient.Discovery()
	// apigroups, err := discovery.GetAPIGroupResources(dd)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// c.restMapper = discovery.NewRESTMapper(apigroups, meta.InterfacesForUnstructured)

	// rest_client, err := c.client.RESTClient()
	// if err != nil {
	// 	return err
	// }
	discoveryCacheDir := filepath.Join("./.kube", "cache", "discovery")
	httpCacheDir := filepath.Join("./.kube", "http-cache")
	discoveryClient, err := discovery.NewCachedDiscoveryClientForConfig(
		config,
		discoveryCacheDir,
		httpCacheDir,
		time.Duration(10*time.Minute))
	// discoveryClient := c.client.DiscoveryClient

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	c.restMapper = expander

	return nil
}

func (c *K8SClient) DeployInstance(name string) error {
	// read template
	// Substitute the variables inside of brackets {{}}
	// create golang object structure
	// deploy the .yml file

	log.Infof("Deployng instance %s", name)

	objects, gvks, err := c.parseTemplateToObjects(name)
	if err != nil {
		return err
	}

	// log.Debugf("Objects %++v", objects)

	// js, _ := json.MarshalIndent(objects, "", " ")
	// log.Debugf(string(js))
	// c.k8sclient.
	for i, obj := range objects {
		gvk := gvks[i]
		// gvr := schema.GroupVersionResource{
		// 	Group: ,
		// 	Version: obj.apiVersion,
		// 	Resource: ,
		// }
		// gvks, _, err := scheme.Scheme.ObjectKinds(obj)
		// if err != nil {
		// 	return err
		// }
		// c.k8sclient.
		// gvk := gvks[0]
		mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		js, _ := json.MarshalIndent(mapping, "", " ")
		log.Debugf("Mapping %s", js)

		asUnstructured := &unstructured.Unstructured{}
		if err := scheme.Scheme.Convert(obj, asUnstructured, nil); err != nil {
			return err
		}

		js, _ = json.MarshalIndent(asUnstructured, "", " ")
		log.Debugf("asUnstructured %s", js)

		// actualObject, err := c.k8sclient.Resource(mapping.Resource).Namespace("").Create(asUnstructured, metav1.CreateOptions{})
		actualObject := c.k8sclient.Resource(mapping.Resource).Namespace(c.config.K8SNamespace)
		// res, err := actualObject.List(metav1.ListOptions{})
		res, err := actualObject.Create(asUnstructured, metav1.CreateOptions{})

		// log.Debugf("%++v", res)
		// js, _ = json.MarshalIndent(res, "", " ")
		// log.Debugf("Operation %s", js)

		if err != nil {
			// log.Fatal(err)
			return err
		}
		// c.client.
		log.Debugf("%++v", obj)
		log.Debugf("%++v", actualObject)
		// c.k8sclient.Resource(obj.)
		// log.Debugf("%++v", c.k8sclient.Resource)
	}

	return nil
}

func (c *K8SClient) parseTemplateToObjects(instanceName string) ([]runtime.Object, []schema.GroupVersionKind, error) {
	template, err := ioutil.ReadFile(c.config.K8STemplate)
	if err != nil {
		return nil, nil, err
	}

	rawSubstitutionMap := c.config.TemplateSubstitutions
	subMap, err := refineMap(rawSubstitutionMap, map[string]interface{}{
		"InstanceName": instanceName,
		"ImageName":    c.config.ImageName,
	})

	if err != nil {
		return nil, nil, err
	}
	// log.Debugf("substitution map %++v", subMap)
	parsed, err := ParseTemplate(string(template), subMap)

	log.Debugf("%s", parsed)

	if err != nil {
		return nil, nil, err
	}

	var objects []runtime.Object
	var gvks []schema.GroupVersionKind
	for _, resource := range strings.Split(parsed, "---") { //
		obj, gvk, err := decode([]byte(resource), nil, nil)
		// obj, _, err := decode([]byte(resource), nil, nil)
		if err != nil {
			// log.Fatal(err)
			// return nil, nil, err
		} else {
			objects = append(objects, obj)
			gvks = append(gvks, *gvk)
		}
	}
	return objects, gvks, nil
}

// func (c *K8SClient) StopInstance(name string) error {
// 	return nil
// }
//
// func (c *K8SClient) DeleteInstance(name string) error {
// 	return nil
// }
//
// func (c *K8SClient) GetInstanceStatus(name string) (model.InstanceStatus, error) {
// 	return nil, nil
// }
//
// func (c *K8SClient) UpdateInstanceInformation(instance *model.Instance) error {
// 	return nil
// }
//
// func (c *K8SClient) GetInstanceUrl(name string) (string, error) {
// 	return "", nil
// }

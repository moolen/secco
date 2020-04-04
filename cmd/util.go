package cmd

// func newClient() (*kubernetes.Clientset, error) {
// 	var cfg *rest.Config
// 	var err error
// 	kubeConfig := viper.GetString("kubeconfig")
// 	if kubeConfig == "" {
// 		cfg, err = rest.InClusterConfig()
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else {
// 		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
// 			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfig}, &clientcmd.ConfigOverrides{}).ClientConfig()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return kubernetes.NewForConfig(cfg)
// }

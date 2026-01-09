package tolliver

//
// import (
// 	"crypto/tls"
// )
//
// func NewInstance(options *InstanceOptions) (Instance, error) {
// 	err := populateDefaults(options)
// 	if err != nil {
// 		return Instance{}, err
// 	}
//
// 	certs := make([]tls.Certificate, 1)
// 	certs[0] = *options.InstanceCert
//
// 	c := Instance{
// 		ConnectionPool:       make([]connectionWrapper, InitialConnectionCapacity),
// 		RetryInterval:        options.RetryInterval,
// 		InstanceCertificates: certs,
// 		ListeningPort:        options.Port,
// 		DatabasePath:         options.DatabasePath,
// 		CA:                   options.CA,
// 	}
//
// 	err = c.initDatabase()
// 	if err != nil {
// 		return Instance{}, err
// 	}
//
// 	for _, v := range options.RemoteAddrs {
// 		c.NewConnection(v)
// 	}
//
// 	if options.Port != -1 {
// 		err = c.listenOn(options.Port)
// 		if err != nil {
// 			return Instance{}, err
// 		}
// 	}
//
// 	return c, nil
// }

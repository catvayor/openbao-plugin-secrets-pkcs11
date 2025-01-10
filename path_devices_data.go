package vaultpkcs11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
	//"github.com/miekg/pkcs11"
)

func (b *backend) pathDevicesData() *framework.Path {
	return &framework.Path{
		Pattern: "devices/" + framework.GenericNameRegex("device_name") + "/$",
		Fields: map[string]*framework.FieldSchema{
			"device_name": &framework.FieldSchema{
				Type:     framework.TypeString,
				Required: true,
				Description: `
	Path to the stored object.`,
			},
		},
		HelpSynopsis:    "List data objects stored in the device root",
		HelpDescription: "List data objects stored in the device root",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.pathDevicesDataRootList,
		},
	}
}
func (b *backend) pathDevicesDataCRUD() *framework.Path {
	return &framework.Path{
		Pattern: "devices/" + framework.GenericNameRegex("device_name") + "/" + framework.MatchAllRegex("path"),

		HelpSynopsis:    "Interact with pkcs11 objects stored in a device",
		HelpDescription: ``,
		Fields: map[string]*framework.FieldSchema{
			"path": &framework.FieldSchema{
				Type:     framework.TypeString,
				Required: true,
				Description: `
	Path to the stored object.`,
			},
			"device_name": &framework.FieldSchema{
				Type:     framework.TypeString,
				Required: true,
				Description: `
	Name of the device this object belongs to.`,
			},
		},
		ExistenceCheck: b.pathDevicesExistenceCheck,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathDevicesDataRead,
			logical.CreateOperation: b.pathDevicesDataWrite,
			logical.UpdateOperation: b.pathDevicesDataWrite,
			logical.DeleteOperation: b.pathDevicesDataDelete,
			logical.ListOperation:   b.pathDevicesDataList,
		},
	}
}

// pathDevicesExistenceCheck is used to check if a given data key exists.
func (b *backend) pathDevicesDataExistenceCheck(ctx context.Context, req *logical.Request, d *framework.FieldData) (bool, error) {
	nameRaw, ok := d.GetOk("device_name")
	if !ok {
		return true, errMissingFields("device_name")
	}
	name := nameRaw.(string)

	pathRaw, ok := d.GetOk("path")
	if !ok {
		return true, errMissingFields("path")
	}
	path := pathRaw.(string)

	key := "devices/" + name + "/" + path

	if k, err := b.GetDevice(ctx, req.Storage, key); err != nil || k == nil {
		return false, nil
	}
	return true, nil
}

// pathDevicesRead corresponds to GET devices/:device_name/:path and is used to get the contents of the data object
func (b *backend) pathDevicesDataRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	nameRaw, ok := d.GetOk("device_name")
	if !ok {
		return nil, errMissingFields("device_name")
	}
	name := nameRaw.(string)

	pathRaw, ok := d.GetOk("path")
	if !ok {
		return nil, errMissingFields("path")
	}
	path := pathRaw.(string)

	// Read the path
	out, err := b.readData(name, path)
	if err != nil {
		return nil, fmt.Errorf("pathDevicesDataRead: read failed: %v", err)
	}

	// Fast-path the no data case
	if out == nil {
		return nil, nil
	}

	// Decode the data
	vData := map[string]interface{}{}
	if err := json.Unmarshal(out, &vData); err != nil {
		return nil, err
	}
	//rawData["Data"] = out
	return &logical.Response{
		//Secret: &logical.Secret{},
		Data: map[string]interface{}{
			"data": vData,
		},
	}, nil

	// var resp *logical.Response
	// if b.generateLeases {
	// 	// Generate the response
	// 	resp = b.Secret("pkcs11").Response(rawData, nil)
	// 	resp.Secret.Renewable = false
	// } else {
	// 	resp = &logical.Response{
	// 		Secret: &logical.Secret{},
	// 		Data:   rawData,
	// 	}
	// }

	// // Check if there is a ttl key
	// ttlDuration := b.System().DefaultLeaseTTL()
	// ttlRaw, ok := rawData["ttl"]
	// if !ok {
	// 	ttlRaw, ok = rawData["lease"]
	// }
	// if ok {
	// 	dur, err := parseutil.ParseDurationSecond(ttlRaw)
	// 	if err == nil {
	// 		ttlDuration = dur
	// 	}

	// 	if b.generateLeases {
	// 		resp.Secret.Renewable = true
	// 	}
	// }

	// resp.Secret.TTL = ttlDuration

	// return resp, nil

}

// pathDevicesList corresponds to LIST devices/:device_name and is used to list all data objects at the root level of the device
func (b *backend) pathDevicesDataRootList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Debug("pathDevicesDataRootList", "FieldData", spew.Sdump(d))
	nameRaw, ok := d.GetOk("device_name")
	if !ok {
		return nil, errMissingFields("device_name")
	}
	name := nameRaw.(string)

	path := "devices/" + name

	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// List the keys at the prefix given by the request
	keys, err := req.Storage.List(ctx, path)
	if err != nil {
		return nil, err
	}

	// Generate the response
	return logical.ListResponse(keys), nil

}

// pathDevicesList corresponds to LIST devices/:device_name/:path and is used to list all data objects in the key
func (b *backend) pathDevicesDataList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Debug("pathDevicesDataList", "FieldData", spew.Sdump(d))
	nameRaw, ok := d.GetOk("device_name")
	if !ok {
		return nil, errMissingFields("device_name")
	}
	name := nameRaw.(string)

	pathRaw, ok := d.GetOk("path")
	if !ok {
		return nil, errMissingFields("path")
	}
	path := "devices/" + name + "/" + pathRaw.(string)

	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// List the keys at the prefix given by the request
	keys, err := req.Storage.List(ctx, path)
	if err != nil {
		return nil, err
	}

	// Generate the response
	return logical.ListResponse(keys), nil

}

// pathKeysWrite corresponds to PUT/POST devices/:device_name/:path and creates a
// new data object in the HSM
func (b *backend) pathDevicesDataWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	nameRaw, ok := d.GetOk("device_name")
	if !ok {
		return nil, errMissingFields("device_name")
	}
	name := nameRaw.(string)

	pathRaw, ok := d.GetOk("path")
	if !ok {
		return nil, errMissingFields("path")
	}
	path := pathRaw.(string)
	// Check that some fields are given
	if len(req.Data) == 0 {
		return logical.ErrorResponse("missing data fields"), nil
	}
	// JSON encode the data
	buf, err := json.Marshal(req.Data)
	if err != nil {
		return nil, fmt.Errorf("json encoding failed: %v", err)
	}
	//if buf, ok := req.Data["secret"]; ok {
	err = b.storeData(name, path, buf)
	if err != nil {
		return nil, fmt.Errorf("pathDevicesDataWrite: failed to write: %v", err)
	}
	//}

	/*


		// Write out a new key
		entry := &logical.StorageEntry{
			Key:   "devices/" + name + "/" + path,
			Value: buf,
		}
		if err := req.Storage.Put(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to write: %v", err)
		}*/

	return nil, nil

}

// pathKeysDelete corresponds to DELETE devices/:device_name/:path and deletes an
// existing data object
func (b *backend) pathDevicesDataDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	nameRaw, ok := d.GetOk("device_name")
	if !ok {
		return nil, errMissingFields("device_name")
	}
	name := nameRaw.(string)

	pathRaw, ok := d.GetOk("path")
	if !ok {
		return nil, errMissingFields("path")
	}
	path := pathRaw.(string)

	key := "devices/" + name + "/" + path

	// Delete the key at the request path
	if err := req.Storage.Delete(ctx, key); err != nil {
		return nil, err
	}
	return nil, nil
}

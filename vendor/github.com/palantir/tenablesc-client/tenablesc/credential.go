package tenablesc

import (
	"fmt"
)

const credentialEndpoint = "/credential"

// Credential is massively pared back from the possible types in https://docs.tenable.com/tenablesc/api/Credential.htm
// This is not wired to directly manage credentials, only to find and delete them.
type Credential struct {
	BaseInfo
	Type      string   `json:"type"`
	CanUse    FakeBool `json:"canUse,omitempty"`
	CanManage FakeBool `json:"canManage,omitempty"`
	Tags      string   `json:"tags,omitempty"`
}

type allCredentialsInternal struct {
	Usable     []*Credential `json:"usable,omitempty" tenable:"recurse"`
	Manageable []*Credential `json:"manageable,omitempty" tenable:"recurse"`
}

func (c *Client) GetAllCredentials() ([]*Credential, error) {
	var creds []*Credential

	allCredentials := &allCredentialsInternal{}

	if _, err := c.getResource(credentialEndpoint, &allCredentials); err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	credMap := make(map[ProbablyString]bool)

	for _, c := range allCredentials.Usable {
		credMap[c.ID] = true
		creds = append(creds, c)
	}
	for _, c := range allCredentials.Manageable {
		if _, exists := credMap[c.ID]; !exists {
			creds = append(creds, c)
			credMap[c.ID] = true
		}
	}

	return creds, nil
}

func (c *Client) GetCredential(id string) (*Credential, error) {
	resp := &Credential{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", credentialEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get credential id %s: %w", id, err)
	}

	return resp, nil
}

func (c *Client) DeleteCredential(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", credentialEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete credential with id %s: %w", id, err)
	}

	return nil
}

// SSHCertificateCredentialUpload https://docs.tenable.com/tenablesc/api/Credential.htm#credential_POST
type SSHCertificateCredentialUpload struct {
	BaseInfo
	Tags                string `json:"tags"`
	Type                string `json:"type"`
	Username            string `json:"username"`
	AuthType            string `json:"authType"`
	PublicKey           string `json:"publicKey"`
	PrivateKey          string `json:"privateKey"`
	PrivilegeEscalation string `json:"privilegeEscalation"`
}

// handleSSHCertCredFileUpload takes a key as a string and uploads it as a new file to tenable.sc.
// it returns the name of the fie that it is stored as on tenable.sc.
// this file name is used as a reference when creating or modifying SSH certificate credentials
func (c *Client) handleSSHCertCredFileUpload(key string) (string, error) {
	keyFile, err := c.UploadFileFromString(key, "key", "")
	if err != nil {
		return "", fmt.Errorf("unable to upload key file: %w", err)
	}
	return keyFile.Filename, nil
}

// DBCredentialUpload https://docs.tenable.com/tenablesc/api/Credential.htm#credential_POST
type DBCredentialUpload struct {
	BaseInfo
	Tags     string `json:"tags"`
	Type     string `json:"type"`
	AuthType string `json:"authType"`
	Login    string `json:"login"`
	Password string `json:"password"`
	SID      string `json:"sid"`
	DBType   string `json:"dbType"`
	Port     string `json:"port"`
}

const (
	postgresDBType   = "PostgreSQL"
	passwordAuthType = "password"
	databaseType     = "database"
)

// AddPostgresDBCredential creates a new postgres db credential
func (c *Client) AddPostgresDBCredential(credName, username, password, description, dbName, port string) (*Credential, error) {
	newPostgresDBCredential := DBCredentialUpload{
		BaseInfo: BaseInfo{
			Name:        credName,
			Description: description,
		},
		Type:     databaseType,
		AuthType: passwordAuthType,
		Login:    username,
		Password: password,
		SID:      dbName,
		DBType:   postgresDBType,
		Port:     port,
	}

	resp := &Credential{}
	if _, err := c.postResource(credentialEndpoint, newPostgresDBCredential, resp); err != nil {
		return nil, fmt.Errorf("failed to create postgres db credential: %w", err)
	}

	return resp, nil
}

// AddSSHCertificateCredential does the following:
// 1. takes a public key and private key as strings and uploads those to tenable.sc
// 2. for both private and public keys it stores the file name in PublicKey and PrivateKey in SSHCertificateCredentialUpload
// 3. post the cred to tenable.sc and return the object
func (c *Client) AddSSHCertificateCredential(publicKey, privateKey string, sshCertCred SSHCertificateCredentialUpload) (*Credential, error) {
	var err error
	sshCertCred.PrivateKey, err = c.handleSSHCertCredFileUpload(privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to upload private key credential for credential name %s: %w", sshCertCred.Name, err)
	}

	sshCertCred.PublicKey, err = c.handleSSHCertCredFileUpload(publicKey)
	if err != nil {
		return nil, fmt.Errorf("unable to upload public key credential for credential name %s: %w", sshCertCred.Name, err)
	}

	resp := &Credential{}
	if _, err := c.postResource(credentialEndpoint, sshCertCred, resp); err != nil {
		return nil, fmt.Errorf("failed to create ssh cert credential: %w", err)
	}

	return resp, nil
}

// UpdateSSHCertificateCredential does the following:
// 1. takes a public key and private key as strings and uploads those to tenable.sc
// 2. for both private and public keys it stores the file name in PublicKey and PrivateKey in SSHCertificateCredentialUpload
// 3. patches the cred to tenable.sc and return the object
func (c *Client) UpdateSSHCertificateCredential(publicKey, privateKey string, sshCertCred SSHCertificateCredentialUpload) (*Credential, error) {
	var err error
	sshCertCred.PrivateKey, err = c.handleSSHCertCredFileUpload(privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to upload private key credential for credential name %s: %w", sshCertCred.Name, err)
	}

	sshCertCred.PublicKey, err = c.handleSSHCertCredFileUpload(publicKey)
	if err != nil {
		return nil, fmt.Errorf("unable to upload public key credential for credential name %s: %w", sshCertCred.Name, err)
	}

	resp := &Credential{}
	if _, err := c.patchResourceWithID(credentialEndpoint, sshCertCred, resp); err != nil {
		return nil, fmt.Errorf("failed to update ssh cert credential: %w", err)
	}

	return resp, nil
}

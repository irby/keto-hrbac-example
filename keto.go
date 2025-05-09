package main

import (
	"context"
	"fmt"
	"log"
	"time"

	rts "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Keto struct {
	checkClient rts.CheckServiceClient
	writeClient rts.WriteServiceClient
}

type KetoConfig struct {
	ReadGPRCAddress  string
	WriteGRPCAddress string
}

func NewKetoClient(ctx context.Context, cfg *KetoConfig) (*Keto, error) {
	readAddress := cfg.ReadGPRCAddress
	writeAddress := cfg.WriteGRPCAddress
	readConn, err := grpc.NewClient(readAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err.Error())
	}
	writeConn, err := grpc.NewClient(writeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err.Error())
	}

	versionClient := rts.NewVersionServiceClient(readConn)

	connected := false
	for i := 0; i < 3; i++ {
		versionResp, err := versionClient.GetVersion(ctx, &rts.GetVersionRequest{})
		if err != nil {
			log.Printf("connection to Keto failed [%s] (%d/3)", err.Error(), i+1)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("connected to Keto (read) server via gRPC at %s [server version %s]", readAddress, versionResp.GetVersion())
		connected = true
		break
	}
	if !connected {
		return nil, fmt.Errorf("failed to connect to Keto at %s", readAddress)
	}

	checkClient := rts.NewCheckServiceClient(readConn)
	writeClient := rts.NewWriteServiceClient(writeConn)
	keto := &Keto{
		checkClient: checkClient,
		writeClient: writeClient,
	}

	return keto, nil
}

func NewSubject(namespace string, subjectId string, relation string) *rts.Subject {
	return &rts.Subject{
		Ref: &rts.Subject_Set{
			Set: &rts.SubjectSet{
				Namespace: namespace,
				Object:    subjectId,
				Relation:  relation,
			},
		},
	}
}

func (k *Keto) CreateRelation(ctx context.Context, namespace string, object string, relation string, subject *rts.Subject) (bool, error) {
	var records []*rts.RelationTuple
	rec := rts.RelationTuple{
		Namespace: namespace,
		Object:    object,
		Relation:  relation,
		Subject:   subject,
	}
	records = append(records, &rec)
	req := rts.TransactRelationTuplesRequest{
		RelationTupleDeltas: rts.RelationTupleToDeltas(records, rts.RelationTupleDelta_ACTION_INSERT),
	}

	_, err := k.writeClient.TransactRelationTuples(ctx, &req)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (k *Keto) DeleteRelation(ctx context.Context, namespace string, object string, relation string, subject *rts.Subject) (bool, error) {
	rec := rts.RelationQuery{
		Namespace: &namespace,
		Object:    &object,
		Relation:  &relation,
		Subject:   subject,
	}
	req := rts.DeleteRelationTuplesRequest{
		RelationQuery: &rec,
	}

	_, err := k.writeClient.DeleteRelationTuples(ctx, &req)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (k *Keto) CheckObjectPermission(ctx context.Context, namespace string, object string, relation string, subject *rts.Subject) (bool, error) {
	req := &rts.CheckRequest{
		Tuple: &rts.RelationTuple{
			Namespace: namespace,
			Object:    object,
			Relation:  relation,
			Subject:   subject,
		},
	}
	res, err := k.checkClient.Check(ctx, req)

	if err != nil {
		return false, err
	}

	return res.Allowed, nil
}

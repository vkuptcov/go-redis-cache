package internal

import (
	"context"

	"github.com/pkg/errors"
)

func HSetKV(ctx context.Context, opts Options, key string, fieldValPairs ...interface{}) error {
	if len(fieldValPairs)%2 != 0 {
		return ErrKeyPairs
	}
	fieldMarshalledValsPairs := make([]interface{}, len(fieldValPairs))
	for idx := 0; idx < len(fieldValPairs); idx += 2 {
		field, ok := fieldValPairs[idx].(string)
		if !ok {
			return errors.Wrapf(ErrNonStringKey, "string field expected for position %d, `%#+v` of type %T given", idx, fieldValPairs[idx], fieldValPairs[idx])
		}
		marshalledBytes, marshalErr := opts.Marshaller.Marshal(fieldValPairs[idx+1])
		if marshalErr != nil {
			return marshalErr
		}
		fieldMarshalledValsPairs[idx] = field
		fieldMarshalledValsPairs[idx+1] = string(marshalledBytes)
	}
	pipeline := opts.Redis.Pipeline()
	pipeline.HSet(ctx, key, fieldMarshalledValsPairs...)
	pipeline.Expire(ctx, key, opts.DefaultTTL)
	_, pipelineErr := pipeline.Exec(ctx)
	return pipelineErr
}

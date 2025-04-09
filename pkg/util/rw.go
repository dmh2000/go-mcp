package util

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"context"
	"encoding/json"
	"log"
	"os"
	mcp "sqirvy/mcp/pkg/mcp"
)

// ReadStdin is no longer used by the server's primary read loop.
// Kept here in case other utilities might use it, but can be removed if unused.
// func ReadStdin(ctx context.Context, logger *log.Logger, msgCh chan<- []byte, errCh chan<- error) {
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			logger.Printf("ReadStdin: %s", scanner.Text())
// 			var req mcp.RPCRequest
// 			if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
// 				errCh <- err
// 				continue
// 			}
// 			// now turn it back into
// 			b, err := json.Marshal(req)
// 			if err != nil {
// 				errCh <- err
// 				continue
// 			}
// 			msgCh <- b
// 			// return // Original bug: returned after first message
// 		}
// 	}
// 	if err := scanner.Err(); err != nil {
// 		errCh <- err
// 	}
// }

func WriteStdout(ctx context.Context, msgCh <-chan mcp.RPCRequest, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logger.Printf("ReadStdin: %s", scanner.Text())
			var req mcp.RPCRequest
			if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
				errCh <- err
				continue
			}
			// now turn it back into
			b, err := json.Marshal(req)
			if err != nil {
				errCh <- err
				continue
			}
			msgCh <- b
			return
		}
	}
	if err := scanner.Err(); err != nil {
		errCh <- err
	}
}

func WriteStdout(ctx context.Context, msgCh <-chan mcp.RPCRequest, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-msgCh:
			data, err := json.Marshal(req)
			if err != nil {
				errCh <- err
				continue
			}
			if _, err := os.Stdout.Write(data); err != nil {
				errCh <- err
				continue
			}
			if _, err := os.Stdout.Write([]byte("\n")); err != nil {
				errCh <- err
				continue
			}
		}
	}
}

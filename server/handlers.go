package server

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Layr-Labs/eigenda-proxy/commitments"
	"github.com/Layr-Labs/eigenda-proxy/store"
	"github.com/gorilla/mux"
)

func (svr *Server) handleHealth(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

// =================================================================================================
// GET ROUTES
// =================================================================================================

// handleGetSimpleCommitment handles the GET request for simple commitments.
func (svr *Server) handleGetSimpleCommitment(w http.ResponseWriter, r *http.Request) error {
	versionByte, err := parseVersionByte(w, r)
	if err != nil {
		return fmt.Errorf("error parsing version byte: %w", err)
	}
	commitmentMeta := commitments.CommitmentMeta{
		Mode:        commitments.SimpleCommitmentMode,
		CertVersion: versionByte,
	}

	rawCommitmentHex, ok := mux.Vars(r)[routingVarNameRawCommitmentHex]
	if !ok {
		return fmt.Errorf("commitment not found in path: %s", r.URL.Path)
	}
	commitment, err := hex.DecodeString(rawCommitmentHex)
	if err != nil {
		return fmt.Errorf("failed to decode commitment %s: %w", rawCommitmentHex, err)
	}

	return svr.handleGetShared(r.Context(), w, commitment, commitmentMeta)
}

// handleGetOPKeccakCommitment handles the GET request for optimism keccak commitments.
func (svr *Server) handleGetOPKeccakCommitment(w http.ResponseWriter, r *http.Request) error {
	// TODO: do we use a version byte in OPKeccak commitments? README seems to say so, but server_test didn't
	// versionByte, err := parseVersionByte(r)
	// if err != nil {
	// 	err = fmt.Errorf("error parsing version byte: %w", err)
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return err
	// }
	commitmentMeta := commitments.CommitmentMeta{
		Mode:        commitments.OptimismKeccak,
		CertVersion: byte(commitments.CertV0),
	}

	rawCommitmentHex, ok := mux.Vars(r)[routingVarNameRawCommitmentHex]
	if !ok {
		return fmt.Errorf("commitment not found in path: %s", r.URL.Path)
	}
	commitment, err := hex.DecodeString(rawCommitmentHex)
	if err != nil {
		return fmt.Errorf("failed to decode commitment %s: %w", rawCommitmentHex, err)
	}

	return svr.handleGetShared(r.Context(), w, commitment, commitmentMeta)
}

// handleGetOPGenericCommitment handles the GET request for optimism generic commitments.
func (svr *Server) handleGetOPGenericCommitment(w http.ResponseWriter, r *http.Request) error {
	versionByte, err := parseVersionByte(w, r)
	if err != nil {
		return fmt.Errorf("error parsing version byte: %w", err)
	}
	commitmentMeta := commitments.CommitmentMeta{
		Mode:        commitments.OptimismGeneric,
		CertVersion: versionByte,
	}

	rawCommitmentHex, ok := mux.Vars(r)[routingVarNameRawCommitmentHex]
	if !ok {
		return fmt.Errorf("commitment not found in path: %s", r.URL.Path)
	}
	commitment, err := hex.DecodeString(rawCommitmentHex)
	if err != nil {
		return fmt.Errorf("failed to decode commitment %s: %w", rawCommitmentHex, err)
	}

	return svr.handleGetShared(r.Context(), w, commitment, commitmentMeta)
}

func (svr *Server) handleGetShared(ctx context.Context, w http.ResponseWriter, comm []byte, meta commitments.CommitmentMeta) error {
	commitmentHex := hex.EncodeToString(comm)
	svr.log.Info("Processing GET request", "commitment", commitmentHex, "commitmentMeta", meta)
	input, err := svr.sm.Get(ctx, comm, meta.Mode)
	if err != nil {
		err = MetaError{
			Err:  fmt.Errorf("get request failed with commitment %v: %w", commitmentHex, err),
			Meta: meta,
		}
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return err
	}

	svr.writeResponse(w, input)
	return nil
}

// =================================================================================================
// POST ROUTES
// =================================================================================================

// handlePostSimpleCommitment handles the POST request for simple commitments.
func (svr *Server) handlePostSimpleCommitment(w http.ResponseWriter, r *http.Request) error {
	commitmentMeta := commitments.CommitmentMeta{
		Mode:        commitments.SimpleCommitmentMode,
		CertVersion: byte(commitments.CertV0), // TODO: hardcoded for now
	}
	return svr.handlePostShared(w, r, nil, commitmentMeta)
}

// handlePostOPKeccakCommitment handles the POST request for optimism keccak commitments.
func (svr *Server) handlePostOPKeccakCommitment(w http.ResponseWriter, r *http.Request) error {
	// TODO: do we use a version byte in OPKeccak commitments? README seems to say so, but server_test didn't
	// versionByte, err := parseVersionByte(r)
	// if err != nil {
	// 	err = fmt.Errorf("error parsing version byte: %w", err)
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return err
	// }
	commitmentMeta := commitments.CommitmentMeta{
		Mode:        commitments.OptimismKeccak,
		CertVersion: byte(commitments.CertV0),
	}

	rawCommitmentHex, ok := mux.Vars(r)[routingVarNameRawCommitmentHex]
	if !ok {
		return fmt.Errorf("commitment not found in path: %s", r.URL.Path)
	}
	commitment, err := hex.DecodeString(rawCommitmentHex)
	if err != nil {
		return fmt.Errorf("failed to decode commitment %s: %w", rawCommitmentHex, err)
	}

	return svr.handlePostShared(w, r, commitment, commitmentMeta)
}

// handlePostOPGenericCommitment handles the POST request for optimism generic commitments.
func (svr *Server) handlePostOPGenericCommitment(w http.ResponseWriter, r *http.Request) error {
	commitmentMeta := commitments.CommitmentMeta{
		Mode:        commitments.OptimismGeneric,
		CertVersion: byte(commitments.CertV0), // TODO: hardcoded for now
	}
	return svr.handlePostShared(w, r, nil, commitmentMeta)
}

func (svr *Server) handlePostShared(w http.ResponseWriter, r *http.Request, comm []byte, meta commitments.CommitmentMeta) error {
	svr.log.Info("Processing POST request", "commitment", hex.EncodeToString(comm), "commitmentMeta", meta)
	input, err := io.ReadAll(r.Body)
	if err != nil {
		err = MetaError{
			Err:  fmt.Errorf("failed to read request body: %w", err),
			Meta: meta,
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	commitment, err := svr.sm.Put(r.Context(), meta.Mode, comm, input)
	if err != nil {
		err = MetaError{
			Err:  fmt.Errorf("put request failed with commitment %v (commitment mode %v): %w", comm, meta.Mode, err),
			Meta: meta,
		}
		if errors.Is(err, store.ErrEigenDAOversizedBlob) || errors.Is(err, store.ErrProxyOversizedBlob) {
			// we add here any error that should be returned as a 400 instead of a 500.
			// currently only includes oversized blob requests
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return err
	}

	responseCommit, err := commitments.EncodeCommitment(commitment, meta.Mode)
	if err != nil {
		err = MetaError{
			Err:  fmt.Errorf("failed to encode commitment %v (commitment mode %v): %w", commitment, meta.Mode, err),
			Meta: meta,
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	svr.log.Info(fmt.Sprintf("response commitment: %x\n", responseCommit))
	// write commitment to resp body if not in OptimismKeccak mode
	if meta.Mode != commitments.OptimismKeccak {
		svr.writeResponse(w, responseCommit)
	}
	return nil
}

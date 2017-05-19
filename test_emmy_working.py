"""
Uses pytest to check that the command line inputs to emmy work as explained in
the README.


Installation on raspberry pi:
(assuming go and emmy are already installed and emmy is in the $PATH)
sudo apt install python3-pytest

Usage:
Move to the directory containing this file
py.test-3

"""

import pytest
import subprocess
import sys
import time


# Subproces has different methods for python2
assert sys.version_info >= (3, )


def run_client_server_test(example, expectedClientError, expectedServerError,
                           expectedServerOutput=b"", expectedClientOutput=b""):
    """Spawns two emmy processes: one server then one client that are each
    running the given example.

    All expected outputs and errors should be passed as a bytestring.

    Tests that expected outputs are INCLUDED in the actual output.
    Tests that the expected errors are EQUAL to the actual error.
    """
    server = subprocess.Popen(["emmy", "-example=" + example, "-client=false"],
                              stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    # Although it would be possible to wait for some kind of output here for
    # some tests, in other tests (schnorr_ec) the server gives no output
    # and therefore no indication that it has started up. Instead, crudely
    # assume that no server will take a whole second to start.
    time.sleep(1)
    client = subprocess.Popen(["emmy", "-example=" + example, "-client=true"],
                              stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    try:
        try:
            clientOut, clientError = client.communicate()
            print(clientError, clientOut)
        except:
            client.kill()
            raise
        assert clientOut == expectedClientOutput
        assert expectedClientError in clientError
    finally:
        server.kill()
    serverOut, serverError = server.communicate()
    print(serverError, serverOut)
    assert serverOut == expectedServerOutput
    assert expectedServerError in serverError


def test_go_tests():
    """Runs their go tests - necessary to run this before cspaillier as this is
    currently (commit d80649a) how the pub/sec keys are created."""
    subprocess.check_output(["go", "test", "github.com/xlab-si/emmy/test"])


def test_schnorr():
    """
    Tests schnorr sigma protocol
    emmy -example=schnorr -client=false
    emmy -example=schnorr -client=true
    """
    run_client_server_test("schnorr", b"knowledge proved", b"1")


def test_schnorr_zkp():
    """
    Tests schnorr zkp
    emmy -example=schnorr_zkp -client=false
    emmy -example=schnorr_zkp -client=true
    """
    run_client_server_test("schnorr_zkp", b"knowledge proved", b"2")


def test_schnorr_zkpok():
    """
    Tests schnorr zkpok
    emmy -example=schnorr_zkpok -client=false
    emmy -example=schnorr_zkpok -client=true
    """
    run_client_server_test("schnorr_zkpok", b"knowledge proved", b"3")


def test_schnorr_ec():
    """
    Tests schnorr ec
    emmy -example=schnorr_ec -client=false
    emmy -example=schnorr_ec -client=true
    """
    run_client_server_test("schnorr_ec", b"proved", b"")


def test_schnorr_ec_zkp():
    """
    Tests schnorr ec_zkp
    emmy -example=schnorr_ec_zkp -client=false
    emmy -example=schnorr_ec_zkp -client=true
    """
    run_client_server_test("schnorr_ec_zkp", b"proved", b"")


def test_schnorr_ec_zkpok():
    """
    Tests schnorr ec_zpkok
    emmy -example=schnorr_ec_zkpok -client=false
    emmy -example=schnorr_ec_zkpok -client=true
    """
    run_client_server_test("schnorr_ec_zkpok", b"proved", b"")


def test_dlog_equality():
    """
    Tests Chaum-Pedersen protocol to prove discrete logarithm equality
    emmy -example=dlog_equality
    """
    out, err = subprocess.Popen(["emmy", "-example=dlog_equality"],
                                stdout=subprocess.PIPE, stderr=subprocess.PIPE
                                ).communicate()
    assert b"true" in err
    assert b"" == out


def test_dlog_equality_blinded_transcript():
    """
    Tests discrete logarithm equality that produces a blinded transcript
    emmy -example=dlog_equality_blinded_transcript
    """
    cmd = ["emmy", "-example=dlog_equality_blinded_transcript"]
    out, err = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE).communicate()
    assert out == b""
    linesOfError = err.decode().splitlines()
    assert "true" in linesOfError[0]
    assert "is the transcript valid" in linesOfError[1]
    assert "true" in linesOfError[2]


def test_pedersen():
    """
    Tests a pedersen commitment
    emmy -example=pedersen -client=false
    emmy -example=pedersen -client=true
    """
    run_client_server_test("pedersen", b"decommitting", b"",
                           expectedClientOutput=b"ok\n")


def test_pedersen_ec():
    """
    Tests a pedersen commitment
    emmy -example=pedersen_ec -client=false
    emmy -example=pedersen_ec -client=true
    """
    run_client_server_test("pedersen_ec", b"decommitting", b"",
                           expectedClientOutput=b"ok\n")


def test_pseudonymsys():
    """
    Tests pseudonym system
    emmy -example=pseudonymsys
    """
    out, err = subprocess.Popen(["emmy", "-example=pseudonymsys"],
                                stdout=subprocess.PIPE, stderr=subprocess.PIPE
                                ).communicate()
    assert b"" == out
    assert b"true" in err


def test_pedersen_ec():
    """
    emmy -example=cspaillier -client=false
    """
    run_client_server_test("cspaillier", b"proved", b"<nil>")


def test_split_secret():
    """
    Tests shamir's secret sharing scheme
    emmy -example=split_secret -client=false
    """
    out, err = subprocess.Popen(["emmy", "-example=split_secret",
                                 "-client=false"], stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE
                                ).communicate()
    assert b"" == out
    assert b"password" in err

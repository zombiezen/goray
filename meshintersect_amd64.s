// intersect_amd64.s
// vim: ft=9asm ai

/*
    This makes heavy use of Intel's SSE2, which is available on P4 and later.
    Every XMM register holds two float64 values and we can do vertical
    operations against two registers.

    Note: The MOVUPD instruction reverses the doubles from memory layout-- A, B
    loads such that B is Hi and A is Lo.  Yes, this makes everything confusing.
    Luckily, the reversal also happens when you go to write the values back to
    memory.

    Also Note: GDB flips the XMM register readout (i.e. Lo, Hi).

    Step 1: edge1, edge2 = b - a, c - a
        X0 |by bx| X1 |cx bz| X2 |cz cy|
        X3 |ay ax| X4 |ax az| X5 |az ay|
    Cross product:
        For each component, use two XMM registers to calculate the determinant.
        For the Y component, we simply swap the columns to negate the value.

        X0 |a b|
        X1 |d c|

        1) Multiply X1 by X0, store into X0
        2) Copy X0's lower value into X1's higher value (using shuffle)
        3) Subtract X1 from X0, store into X0
        4) The high value from X0 is the result.
    Dot product:
        This is another weird one, motivated by a desire to have the result in a
        low register so that we can immediately multiply by other scalars.

        X0 |ay ax| X1 |0 az|
        X2 |by bx| X3 |0 bz|

        1) Multiply X2 by X0, store into X0
        2) Multiply X3 by X1, store into X1
        3) Copy X0 into X2
        4) Swap high and low in X2
        5) Add X2 to X0, store into X0.
        6) Add X1 to X0, store into X0.
        7) The low (scalar) value from X0 is the result.
 */

// func intersect(a, b, c, rDir, rFrom vec64.Vector) (rayDepth, u, v float64)
// a: +0(FP)
// b: +24(FP)
// c: +48(FP)
// rDir: +72(FP)
// rFrom: +96(FP)
//
// Return values:
// rayDepth: +120(FP)
// u: +128(FP)
// v: +136(FP)
//
// Locals:
// edge2: +-24(SP)
// edge1: +-48(SP)
// pvec/qvec: +-72(SP)
// tvec: +-96(SP)
//
// X5: u
// X6: v
// X7: d
TEXT ·intersect(SB),7,$96-144
    // Step 1: edge1, edge2 = b - a, c - a
    // Move vertices into XMM registers
    MOVUPD      bXY+24(FP), X0
    MOVUPD      bZcX+40(FP), X1
    MOVUPD      cYZ+56(FP), X2
    MOVUPD      aXY+0(FP), X3
    MOVHPD      aX+0(FP), X4
    MOVLPD      aZ+16(FP), X4
    MOVUPD      aYZ+8(FP), X5
    // Perform subtraction
    SUBPD       X3, X0
    SUBPD       X4, X1
    SUBPD       X5, X2
    // Store edges
    MOVUPD      X0, edge1XY+-48(SP)
    MOVUPD      X1, edge1Z2X+-32(SP)
    MOVUPD      X2, edge2YZ+-16(SP)
    // Step 2: pvec = rDir cross edge2
    // X
    MOVUPD      rDirYZ+80(FP), X0
    SHUFPD      $1, X0, X0
    MOVUPD      edge2YZ+-16(SP), X1
    MULPD       X1, X0
    SHUFPD      $0, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, pvecX+-72(SP)
    // Y
    MOVHPD      rDirZ+88(FP), X0
    MOVLPD      rDirX+72(FP), X0
    MOVHPD      edge2X+-24(SP), X1
    MOVLPD      edge2Z+-8(SP), X1
    MULPD       X1, X0
    SHUFPD      $0, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, pvecY+-64(SP)
    // Z
    MOVUPD      rDirXY+72(FP), X0
    SHUFPD      $1, X0, X0
    MOVUPD      edge2XY+-24(SP), X1
    MULPD       X1, X0
    SHUFPD      $0, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, pvecZ+-56(SP)
    // Step 3: d = edge1 dot pvec
    MOVUPD      edge1XY+-48(SP), X0
    MOVUPD      pvecXY+-72(SP), X2
    MULPD       X2, X0
    MOVUPD      X0, X2
    SHUFPD      $1, X2, X2
    MOVSD       edge1Z+-32(SP), X1
    MOVSD       pvecZ+-56(SP), X3
    MULSD       X3, X1
    ADDSD       X2, X0
    ADDSD       X1, X0
    // Step 4: If d == 0, no collision.
    // (d is already in register X0)
    MOVSD       $0.0, X1
    UCOMISD     X0, X1
    JE          NoCollision
    // Step 5: d = 1.0 / d
    // (d is already in register X0)
    MOVSD       $1.0, X7
    DIVSD       X0, X7
    // Step 6: tvec = rFrom - a
    MOVUPD      rFromXY+96(FP), X0
    MOVUPD      aXY+0(FP), X2
    SUBPD       X2, X0
    MOVUPD      X0, tvecXY+-96(SP)
    MOVSD       rFromZ+112(FP), X1
    MOVSD       aZ+16(FP), X3
    SUBSD       X3, X1
    MOVSD       X1, tvecZ+-80(SP)
    // Step 7: u = (pvec dot tvec) * d
    MOVUPD      pvecXY+-72(SP), X5
    MOVUPD      tvecXY+-96(SP), X2
    MULPD       X2, X5
    MOVUPD      X5, X2
    SHUFPD      $1, X2, X2
    MOVSD       pvecZ+-56(SP), X1
    MOVSD       tvecZ+-80(SP), X3
    MULSD       X3, X1
    ADDSD       X2, X5
    ADDSD       X1, X5
    MULSD       X7, X5
    // Step 8: if u < 0 || u > 1 { return }
    MOVSD       $0.0, X1
    UCOMISD     X1, X5
    JB          NoCollision
    MOVSD       $1.0, X1
    UCOMISD     X1, X5
    JA          NoCollision
    // Step 9: qvec = tvec cross edge1
    // X
    MOVUPD      tvecYZ+-88(SP), X0
    SHUFPD      $1, X0, X0
    MOVUPD      edge1Y+-40(SP), X1
    MULPD       X1, X0
    SHUFPD      $0, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, qvecX+-72(SP)
    // Y
    MOVHPD      tvecZ+-80(SP), X0
    MOVLPD      tvecX+-96(SP), X0
    MOVHPD      edge1X+-48(SP), X1
    MOVLPD      edge1Z+-32(SP), X1
    MULPD       X1, X0
    SHUFPD      $0, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, qvecY+-64(SP)
    // Z
    MOVUPD      tvecXY+-96(SP), X0
    SHUFPD      $1, X0, X0
    MOVUPD      edge1XY+-48(SP), X1
    MULPD       X1, X0
    SHUFPD      $0, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, qvecZ+-56(SP)
    // Step 10: v = (rDir dot qvec) * d
    MOVUPD      rDirXY+72(FP), X6
    MOVUPD      qvecXY+-72(SP), X2
    MULPD       X2, X6
    MOVUPD      X6, X2
    SHUFPD      $1, X2, X2
    MOVSD       rDirZ+88(FP), X1
    MOVSD       qvecZ+-56(SP), X3
    MULSD       X3, X1
    ADDSD       X2, X6
    ADDSD       X1, X6
    MULSD       X7, X6
    // Step 11: if v < 0 || u + v > 1 { return }
    // (v is already in register X0)
    MOVSD       $0.0, X1
    UCOMISD     X1, X6
    JB          NoCollision
    MOVSD       X5, X0
    ADDSD       X6, X0
    MOVSD       $1.0, X1
    UCOMISD     X1, X0
    JA          NoCollision
    // Step 12: rayDepth = (edge2 dot qvec)
    MOVUPD      edge2XY+-24(SP), X0
    MOVUPD      qvecXY+-72(SP), X2
    MULPD       X2, X0
    MOVUPD      X0, X2
    SHUFPD      $1, X2, X2
    MOVSD       edge2Z+-8(SP), X1
    MOVSD       qvecZ+-56(SP), X3
    MULSD       X3, X1
    ADDSD       X2, X0
    ADDSD       X1, X0
    MULSD       X7, X0
    MOVSD       X0, rayDepth+120(FP)
    // Step 13: return
    MOVSD       X5, u+128(FP)
    MOVSD       X6, v+136(FP)
    RET
    // No collision return
NoCollision:
    MOVSD       $-1.0, X0
    MOVSD       X0, rayDepth+120(FP)
    RET

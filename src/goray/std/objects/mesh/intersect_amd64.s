// intersect_amd64.s
// vim: ft=9asm ai

/*
    This makes heavy use of Intel's SSE2, which is available on P4 and later.
    Every XMM register holds two float64 values and we can do vertical
    operations against two registers.

    Vector addition/subtraction:
        We can simply store components in order and operate. In Step 1, we
        subtract two vectors simultaneously.
    Cross product:
        For each component, use two XMM registers to calculate the
        determinant.  Reverse the bottom row like so:

        X0 | a b |
        X1 | d c |

        1) Multiply X1 by X0, store into X0
        2) Copy X0's lower value into X1's higher value (using shuffle)
        3) Subtract X1 from X0, store into X0
        4) The high value from X0 is the result.
    Dot product:
 */

// func intersect(a, b, c, rDir, rFrom [3]float64) (rayDepth, u, v float64)
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
// edge1: +0(SP)
// edge2: +24(SP)
// pvec/qvec: +48(SP)
// tvec: +72(SP)
// d: +96(SP)
TEXT ·intersect(SB),7,$104
    // Step 1: edge1, edge2 = b - a, c - a
    // Move vertices into XMM registers
    MOVUPD      bXY+24(FP), X0
    MOVUPD      bZcX+40(FP), X1
    MOVUPD      cYZ+56(FP), X2
    MOVUPD      aXY+0(FP), X3
    MOVHPD      aZ+16(FP), X4
    MOVLPD      aX+0(FP), X4
    MOVUPD      aYZ+8(FP), X5
    // Perform subtraction
    SUBPD       X3, X0
    SUBPD       X4, X1
    SUBPD       X5, X2
    // Store edges
    MOVUPD      X0, edge1XY+0(SP)
    MOVUPD      X1, edge1Z2X+16(SP)
    MOVUPD      X2, edge2YZ+32(SP)
    // Step 2: pvec = rDir cross edge2
    // X
    MOVUPD      rDirYZ+80(FP), X0
    MOVHPD      edge2Z+40(SP), X1
    MOVLPD      edge2Y+32(SP), X1
    MULPD       X1, X0
    SHUFPD      $2, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, pvecX+48(SP)
    // Y
    MOVHPD      rDirX+72(FP), X0
    MOVLPD      rDirZ+88(FP), X0
    MOVHPD      edge2Z+40(SP), X1
    MOVLPD      edge2X+24(SP), X1
    MULPD       X1, X0
    SHUFPD      $2, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, pvecY+56(SP)
    // Z
    MOVUPD      rDirXY+72(FP), X0
    MOVHPD      edge2Y+32(SP), X1
    MOVLPD      edge2X+24(SP), X1
    MULPD       X1, X0
    SHUFPD      $2, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, pvecZ+64(SP)
    // Step 3: edge1 dot pvec
    MOVUPD      edge1XY+0(SP), X0
    MOVHPD      edge1Z+16(SP), X1
    MOVUPD      pvecXY+48(SP), X2
    MOVHPD      pvecZ+64(SP), X3
    MULPD       X2, X0
    MULPD       X3, X1
    SHUFPD      $2, X0, X1
    ADDPD       X2, X0
    ADDPD       X1, X0
    MOVHPD      X0, d+96(SP)
    // Step 4: If d == 0, no collision.
    MOVSD       d+96(SP), X0
    MOVSD       $0.0, X1
    UCOMISD     X0, X1
    JE          NoCollision
    // Step 5: d = 1.0 / d
    // (d is already in register X0)
    MOVSD       $1.0, X1
    DIVSD       X0, X1
    MOVSD       X1, d+96(SP)
    // Step 6: tvec = rFrom - a
    MOVUPD      rFromXY+96(FP), X0
    MOVSD       rFromZ+112(FP), X1
    MOVUPD      tvecXY+72(SP), X2
    MOVSD       tvecZ+88(SP), X3
    SUBPD       X2, X0
    SUBPD       X3, X1
    MOVUPD      X0, tvecXY+72(SP)
    MOVSD       X1, tvecZ+88(SP)
    // Step 7: u = (pvec dot tvec) * d
    MOVUPD      pvecXY+48(SP), X0
    MOVHPD      pvecZ+64(SP), X1
    MOVUPD      tvecXY+72(SP), X2
    MOVHPD      tvecZ+88(SP), X3
    MULPD       X2, X0
    MULPD       X3, X1
    SHUFPD      $2, X0, X1
    ADDPD       X2, X0
    ADDPD       X1, X0
    MOVHPD      d+96(SP), X1
    MULPD       X1, X0
    MOVHPD      X0, u+128(FP)
    // Step 8: if u < 0 || u > 1 { return }
    MOVSD       u+128(FP), X0
    MOVSD       $0.0, X1
    UCOMISD     X0, X1
    JNAE        NoCollision
    MOVSD       $1.0, X1
    UCOMISD     X0, X1
    JNBE        NoCollision
    // Step 9: qvec = tvec cross edge1
    // X
    MOVUPD      tvecYZ+80(SP), X0
    MOVHPD      edge1Z+16(SP), X1
    MOVLPD      edge1Y+8(SP), X1
    MULPD       X1, X0
    SHUFPD      $2, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, qvecX+48(SP)
    // Y
    MOVHPD      tvecX+72(SP), X0
    MOVLPD      tvecZ+88(SP), X0
    MOVHPD      edge1Z+16(SP), X1
    MOVLPD      edge1X+0(SP), X1
    MULPD       X1, X0
    SHUFPD      $2, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, qvecY+56(SP)
    // Z
    MOVUPD      tvecXY+72(SP), X0
    MOVHPD      edge1Y+8(SP), X1
    MOVLPD      edge1X+0(SP), X1
    MULPD       X1, X0
    SHUFPD      $2, X0, X1
    SUBPD       X1, X0
    MOVHPD      X0, qvecZ+64(SP)
    // Step 10: v = (rDir dot qvec) * d
    MOVUPD      rDirXY+72(FP), X0
    MOVHPD      rDirZ+88(FP), X1
    MOVUPD      qvecXY+48(SP), X2
    MOVHPD      qvecZ+64(SP), X3
    MULPD       X2, X0
    MULPD       X3, X1
    SHUFPD      $2, X0, X1
    ADDPD       X2, X0
    ADDPD       X1, X0
    MOVHPD      d+96(SP), X1
    MULPD       X1, X0
    MOVHPD      X0, v+136(FP)
    // Step 11: if v < 0 || u + v > 1 { return }
    MOVSD       v+136(FP), X0
    MOVSD       $0.0, X1
    UCOMISD     X0, X1
    JNAE        NoCollision
    MOVSD       u+128(FP), X1
    ADDSD       X1, X0
    MOVSD       $1.0, X1
    UCOMISD     X0, X1
    JNBE        NoCollision
    // Step 12: rayDepth = (edge2 dot qvec)
    MOVUPD      edge2XY+24(SP), X0
    MOVHPD      edge2Z+40(SP), X1
    MOVUPD      qvecXY+48(SP), X2
    MOVHPD      qvecZ+64(SP), X3
    MULPD       X2, X0
    MULPD       X3, X1
    SHUFPD      $2, X0, X1
    ADDPD       X2, X0
    ADDPD       X1, X0
    MOVHPD      X0, rayDepth+120(FP)
    // Step 13: return
    RET
    // No collision return
NoCollision:
    MOVSD       $-1.0, X0
    MOVSD       X0, rayDepth+120(FP)
    RET

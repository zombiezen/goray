// vector/op_amd64.s
// vim: ft=9asm ai

// func Add(v1, v2 Vector3D) Vector3D
// v1: +0(FP)
// v2: +24(FP)
// Return: +48(FP)
TEXT ·Add(SB),$0-72
    MOVUPD      v1+0(FP), X0
    MOVUPD      v2+24(FP), X2
    ADDPD       X2, X0
    MOVUPD      X0, ret+48(FP)
    MOVSD       v1Z+16(FP), X0
    MOVSD       v2Z+40(FP), X2
    ADDSD       X2, X0
    MOVSD       X0, retZ+64(FP)
    RET

// func Sub(v1, v2 Vector3D) Vector3D
// v1: +0(FP)
// v2: +24(FP)
// Return: +48(FP)
TEXT ·Sub(SB),$0-72
    MOVUPD      v1+0(FP), X0
    MOVUPD      v2+24(FP), X2
    SUBPD       X2, X0
    MOVUPD      X0, ret+48(FP)
    MOVSD       v1Z+16(FP), X0
    MOVSD       v2Z+40(FP), X2
    SUBSD       X2, X0
    MOVSD       X0, retZ+64(FP)
    RET

// func Dot(v1, v2 Vector3D) float64
// v1: +0(FP)
// v2: +24(FP)
// Return: +48(FP)
TEXT ·Dot(SB),$0-56
    MOVUPD      v1+0(FP), X0
    MOVUPD      v2+24(FP), X2
    MULPD       X2, X0
    MOVUPD      X0, X2
    SHUFPD      $1, X2, X2
    MOVSD       v1Z+16(FP), X1
    MOVSD       v2Z+40(FP), X3
    MULSD       X3, X1
    ADDSD       X2, X0
    ADDSD       X1, X0
    MOVSD       X0, ret+48(FP)
    RET

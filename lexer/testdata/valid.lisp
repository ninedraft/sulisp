(defun test-function (arg1 arg2)
  (let ((list '(1 2 3))
        (string "Hello, World!")
        (number 42)
        (boolean t)
        (nil-value nil))
    (format t "List: ~a~%" list)
    (format t "String: ~a~%" string)
    (format t "Number: ~a~%" number)
    (format t "Boolean: ~a~%" boolean)
    (format t "Nil value: ~a~%" nil-value)
    (+ arg1 arg2)))

(define-constant PI 3.14159)

(setf variable (+ 10 20))

(if (< variable 50)
    (format t "Variable is less than 50.~%")
    (format t "Variable is greater than or equal to 50.~%"))

(do ((i 0 (+ i 1)))
    ((>= i 10))
  (format t "Iteration: ~a~%" i))
# Simple Example Go Demonization

Это простой пример демонизации программы на языке Go. Классический подход демонизации, в котором используется fork() в Go не работает. (более подробно почему так происходит вы можете прочитать здесь https://habr.com/ru/post/187668/). Для демонизации программы мы будем использовать os/exec.Command(). Демонизируется программа, выполняющая следующую задачу:

"Необходимо написать приложение, которое запускается в daemon-режиме и слушает 3000 порт (HTTP). 
Любые запросы обрабатываются одним обработчиком, который используя in-memory хранилище формирут
ответ из списка IP:PORT предыдущих вызовов этого обработчика (включая текущий) с указанием времени обращения. 
Также каждую минуту необходимо вычищать этот список. Само приложение находится в win.go. Демон в demonize.go"

Пример работы такого приложения: http://144.76.60.45:3000/


This is a simple example of demonizing a program on Golang. We can not use fork syscall in Golang's runtime, 
because child process doesn't inherit threads and goroutines in that case. We will use os/exec.Command() function 
to launch a child process of the parent program but with a different parameter. Typical Unix/Linux daemon process 
supports start and stop parameters. In our program start block, it will launch a daemon process but with main 
parameter instead.

After building "demonize.go", use "./demonize start" to run daemom or "./daemonize stop" to stop daemon.

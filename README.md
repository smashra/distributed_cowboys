# Distributed Cowboys

## Environment
Minikube on Mac has been used for simulating a distributed environment. Docker containers represent cowboys. Shooting activity is abstracted by using a messaging platform. Each shooter subscribes and publishes to a topic on the messaging server. 
Orchestration: https://minikube.sigs.k8s.io/docs/
Messaging: https://nats.io/

**Note**: Please consult the document supplied to you to understand the requirements of the task at hand.  

## Implementation details
Shots fired among cowboys are represented as messages published on a topic each participating cowboy is listening to. If a message contains a cowboy's name and the message type is 'SHOT' then the subscribing cowboy must consume that message and update his health information by decrementing it  by the damage amount. Immediately after that the shot cowboy should broadcast his updated health to rest of the shooters. Message type for such a message is 'HEALTH'.  These messages are published on the same topic. Since the shoutout has to start at the same time in parallel, message type 'START' is sent only once at the beginning to commence the shooting.
By the end of the shootout there may be a single winner.
## Setup
For setting up minikube, please follow the instructions here https://minikube.sigs.k8s.io/docs/start/ .
 In my case I have used the docker driver for minikube, it's the default.

Start minikube: 

    minikube start

Enable registry addon for pushing docker images locally

    minikube addons registry enable

Open a terminal and run the following command. This allows us to push docker images locally from host to the minikube's registry.

    sudo docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000

Open another terminal. Install nats on minikube.

    kubectl apply -f https://raw.githubusercontent.com/nats-io/k8s/master/nats-server/single-server-nats.yml

After a successful installation of nats, check the service and the pods using:

    kubectl get svc nats
    kubectl get pods nats-0

Clone the repository :

     git clone https://github.com/smashra/distributed_cowboys.git

Setup the shooters:

    cd distributed_cowboys
    ./bin/setup

Once the setup completes successfully, make sure all the pods have a status of `Running`:

    kubectl get po 

 Clone kubetail for tailing logs from all the pods at the same time:

    git clone https://github.com/johanhaleby/kubetail.git

 Open two terminals, in the first one:
 **Note**: Path to kubetail may vary depending on your install and where it's called from, please check.

    ../kubetail/kubetail bill,john,sam,philip,peter

in the second terminal:

    kubectl exec -it bs -- go run bs.go

Assuming everything went fine, we should see output like below in the first terminal:
**Note**: Actual output is coloured with different colours for different pod logs.

    Using regex '.*bill.*|.*john.*|.*sam.*|.*philip.*|.*peter.*' to match pods
    Will tail 5 logs...
    bill
    john
    peter
    philip
    sam
    [philip] Philip starts shooting ....
    [john] John starts shooting ....
    [peter] Peter starts shooting ....
    [sam] Sam starts shooting ....
    [bill] Bill starts shooting ....
    [peter] Peter shoots John ...
    [sam] Sam shoots Bill ...
    [john] John got shot, health [9]
    [bill] Bill shoots Peter ...
    [bill] Bill got shot, health [6]
    [sam] Sam got shot, health [9]
    [john] John shoots Sam ...
    [philip] Philip shoots Peter ...
    [peter] Peter got shot, health [2]
    [peter] Peter got shot, health [-1]
    [peter] Peter is dead.
    [sam] Sam shoots Philip ...
    [philip] Philip got shot, health [14]
    [john] John shoots Philip ...
    [philip] Philip got shot, health [13]
    [sam] Sam shoots Philip ...
    [philip] Philip got shot, health [12]
    [bill] Bill shoots John ...
    [john] John got shot, health [8]







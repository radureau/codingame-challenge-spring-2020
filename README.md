# Work in progress

### 'food for thoughts'
* https://en.wikipedia.org/wiki/Biconnected_component
* https://en.wikipedia.org/wiki/Iterative_deepening_depth-first_search
* https://en.wikipedia.org/wiki/Depth-first_search#Vertex_orderings

### Decision making

- #Node = #Cell - #Wall = #Space
- #Pac0 = #Ally0 * 2
- #Pellet0 = #Node - #Pac0 - #Cherry0
- #ScorePt0 = #Pellet0 + #Cherry0*10
- ScoreTarget = #ScorePt0/2 +1
- MyProgress = MyScore*100/ScoreTarget
- OpntProgress = OpntScore*100/ScoreTarget
- GameProgress = Game.Turn*100/200 = Game.Turn/2

### Before hand

- [x] Build Graph -> Invariable  
- [x] Map Node Distances -> Invariable  
- [x] Map Node with reached Nodes given a time and velocity -> Invariable  
- [ ] Find cut vertices -> Invariable  
- [x] Initialize Fog of War (track pellets/cherries status and opponents potential positions) -> Variable  

### INIT node with pellet ? score

1st turn:
> set pellet presence score proportional to distance from ally / foe
Each turn:
> decrease pellet presence score
> If pellet is closer to a visible opponent then decrease even more
> If uncovered node with pellet then increase value to path filled with (possible) pellets linked to it
> If uncovered empty node then decrease value to path filled with possible pellets linked to it


### Attribute Score to a move (within an ally influence)
Given  
- _a Fog of War state_
- _MyProgress_
- _OpntProgress_
- _GameProgress_

**[go over](https://en.wikipedia.org/wiki/Iterative_deepening_depth-first_search) next possibilities and score them with these indicators:**
- [x] Any cherry left ?  
- [ ] How many pellet can I eat  
- [ ] How many pellet can I discover  
- [ ] How many ally can I encounter  
- [ ] How many opponents can I encounter
- [ ] How much threat 
- [ ] How much fog of war can I decrease  


### Decide between SWITCH SPEED MOVE
```
If Cooldown > 0 then MOVE
If Turn == 1 || Turn == 2
    If Cherry not next to me then check for opponent or else SPEED
    Else Move to Cherry
If Opponent Next to me
    If I can beat him then move to him
    If he is faster then SWICH (if I can block him do that instead)
    If he can't transform then SWITCH
    If lot of pellet in sight then SPEED
    Else SWITCH
If must/shoud collide with opponent then SWITCH to type beating nearest threat
If MyProgress near end then MOVE
If Lot of pellets left in fog of war AND oponents last seen type not threatning then SPEED
If Can kill oponent Pac then use SWITCH/SPEED/MOVE accordingly
DEFAULT MOVE
```

package src.labs.infexf.agents;

// SYSTEM IMPORTS
import java.util.*; 
import edu.bu.labs.infexf.agents.SpecOpsAgent;
import edu.bu.labs.infexf.distance.DistanceMetric;
import edu.bu.labs.infexf.graph.Vertex;
import edu.bu.labs.infexf.graph.Path;


import edu.cwru.sepia.environment.model.state.Unit.UnitView;
import edu.cwru.sepia.environment.model.state.State.StateView;


// JAVA PROJECT IMPORTS


public class InfilExfilAgent
    extends SpecOpsAgent
{

    public InfilExfilAgent(int playerNum)
    {
        super(playerNum);
    }

    // if you want to get attack-radius of an enemy, you can do so through the enemy unit's UnitView
    // Every unit is constructed from an xml schema for that unit's type.
    // We can lookup the "range" of the unit using the following line of code (assuming we know the id):
    //     int attackRadius = state.getUnit(enemyUnitID).getTemplateView().getRange();
    @Override
    public float getEdgeWeight(Vertex src,
                               Vertex dst,
                               StateView state)
    {
        float edgeWeight = 1f; // Default edge weight

        for (Integer enemyID : getOtherEnemyUnitIDs()) {
            UnitView enemy = state.getUnit(enemyID);
            if (enemy != null) {
                int enemyX = enemy.getXPosition();
                int enemyY = enemy.getYPosition();
                Vertex enemyPos = new Vertex(enemyX,enemyY);

                float distanceToMe = DistanceMetric.chebyshevDistance(src, enemyPos);
                edgeWeight = edgeWeight + (float)Math.pow(1000/distanceToMe, 2); // Inversely proportional edge weight
            }
        }
        return edgeWeight;
    }

    @Override
    public boolean shouldReplacePlan(StateView state)
    {
        // Hint: Only replan when necessary
        Stack<Vertex> currentPlan = getCurrentPlan();

        if (currentPlan == null || currentPlan.isEmpty()) {
            return true;
        }

         for (Vertex myPos : currentPlan) {
            int x = myPos.getXCoordinate();
            int y = myPos.getYCoordinate();
        
            for (Integer enemyID : getOtherEnemyUnitIDs()) {
                UnitView enemyUnit = state.getUnit(enemyID);
                if(enemyUnit != null) {
                    int enemyX = enemyUnit.getXPosition();
                    int enemyY = enemyUnit.getYPosition();
                    Vertex enemyPos = new Vertex(enemyX,enemyY);

                    float distanceToEnemy = DistanceMetric.chebyshevDistance(myPos, enemyPos);

                    if(distanceToEnemy <= state.getUnit(enemyID).getTemplateView().getRange() + 4) {
                        return true;
                    }
                }
            }
        }
        return false;
    }
}

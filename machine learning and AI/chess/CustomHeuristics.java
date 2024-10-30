
	/**
	 * TODO: implement me! The heuristics that I wrote are useful, but not very good for a good chessbot.
	 * Please use this class to add your heuristics here! I recommend taking a look at the ones I provided for you
	 * in DefaultHeuristics.java (which is in the same directory as this file)
	 */

	 package src.pas.chess.heuristics;

	 // SYSTEM IMPORTS
	 import edu.bu.chess.search.DFSTreeNode;
	 import edu.cwru.sepia.util.Direction;
	 import edu.bu.chess.game.move.PromotePawnMove;
	 import edu.bu.chess.game.piece.Piece;
	 import edu.bu.chess.game.piece.PieceType;
	 import edu.bu.chess.game.player.Player;
	 import edu.bu.chess.search.DFSTreeNode;
	 import edu.bu.chess.utils.Coordinate;
	 import edu.bu.chess.game.move.Move; // NEW 
	 import edu.bu.chess.game.move.MoveType; // NEW
	 import edu.bu.chess.game.move.MovementMove; // NEW

	 // JAVA PROJECT IMPORTS
	 import src.pas.chess.heuristics.DefaultHeuristics;
	 import java.util.Map; 
	 import java.util.HashMap; 

	 
	public class CustomHeuristics
	 	extends Object 
	{

		// Keeping track of killer moves
		private static Map<String, Move> killerMoves = new HashMap<>();
		private static Map<String, Integer> historyHeuristic = new HashMap<>(); // History heuristic tracker
	
		public static double getMaxPlayerHeuristicValue(DFSTreeNode node) {
			double offenseHeuristicValue = getOffensiveMaxPlayerHeuristicValue(node)
					+ OffensiveHeuristics.getBoardControlValue(node)
					+ OffensiveHeuristics.getHighValuePieceThreatValue(node)
					+ OffensiveHeuristics.getCenterControlValue(node);  // NEW: Center control value
	
			double defenseHeuristicValue = getDefensiveMaxPlayerHeuristicValue(node)
					+ DefensiveHeuristics.getKingSafetyValue(node)
					+ DefensiveHeuristics.getHighValuePieceProtectionValue(node)
					+ DefensiveHeuristics.getPawnStructurePenalty(node);  // NEW: Penalize weak pawn structures
	
			double nonlinearHeuristicValue = getNonlinearPieceCombinationMaxPlayerHeuristicValue(node);
	
			// Incorporate the killer heuristic to prioritize successful moves
			offenseHeuristicValue += applyKillerHeuristic(node);
			defenseHeuristicValue += applyHistoryHeuristic(node);
			
			return offenseHeuristicValue + defenseHeuristicValue + nonlinearHeuristicValue;
		}

		// getting the min and max players 
		public static Player getMaxPlayer(DFSTreeNode node)
		{
			return node.getMaxPlayer();
		}

		public static Player getMinPlayer(DFSTreeNode node)
		{
			return DefaultHeuristics.getMaxPlayer(node).equals(node.getGame().getCurrentPlayer()) ? node.getGame().getOtherPlayer() : node.getGame().getCurrentPlayer();
		}
	
		// Method to prioritize killer moves
		public static double applyKillerHeuristic(DFSTreeNode node) {
			String boardHash = node.getGame().getBoard().hashCode() + "";
			if (killerMoves.containsKey(boardHash)) {
				Move killerMove = killerMoves.get(boardHash);
				if (node.getMove().equals(killerMove)) {
					return 50; // Boost this move if it matches the killer move
				}
			}
			return 0;
		}
	
		// Method to apply the history heuristic (giving weight to historically successful moves)
		public static double applyHistoryHeuristic(DFSTreeNode node) {
			String moveKey = node.getMove().toString();
			if (historyHeuristic.containsKey(moveKey)) {
				return historyHeuristic.get(moveKey) * 10; // Boost based on how often this move was successful
			}
			return 0;
		}
	
		// Offensive heuristic calculations (material and threats)
		public static double getOffensiveMaxPlayerHeuristicValue(DFSTreeNode node) {
			double damageDealtInThisNode = node.getGame().getBoard().getPointsEarned(getMaxPlayer(node));
	
			if (node.getMove().getType() == MoveType.PROMOTEPAWNMOVE) {
				PromotePawnMove promoteMove = (PromotePawnMove) node.getMove();
				damageDealtInThisNode += Piece.getPointValue(promoteMove.getPromotedPieceType());
			}
	
			int numPiecesWeAreThreatening = OffensiveHeuristics.getNumberOfPiecesMaxPlayerIsThreatening(node);
	
			return damageDealtInThisNode + numPiecesWeAreThreatening;
		}
	
		// Defensive heuristic calculations (surrounding pieces, threats to king)
		public static double getDefensiveMaxPlayerHeuristicValue(DFSTreeNode node) {
			int numPiecesAlive = DefensiveHeuristics.getNumberOfMaxPlayersAlivePieces(node);
			int kingSurroundingPiecesValueTotal = DefensiveHeuristics.getClampedPieceValueTotalSurroundingMaxPlayersKing(node);
			int numPiecesThreateningUs = DefensiveHeuristics.getNumberOfPiecesThreateningMaxPlayer(node);
	
			return numPiecesAlive + kingSurroundingPiecesValueTotal + numPiecesThreateningUs;
		}
	
		public static class OffensiveHeuristics extends Object {
	
			public static int getCenterControlValue(DFSTreeNode node) {
				int centerControlValue = 0;
	
				// Define key central squares
				Coordinate[] centralSquares = new Coordinate[]{
					new Coordinate(3, 3), new Coordinate(3, 4), 
					new Coordinate(4, 3), new Coordinate(4, 4)
				};
	
				// Reward for controlling central squares with high-value pieces
				for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					for (Move move : piece.getAllMoves(node.getGame())) {
						if (move instanceof MovementMove) {
							MovementMove movementMove = (MovementMove) move;
							Coordinate target = movementMove.getTargetPosition();
							for (Coordinate center : centralSquares) {
								if (target.equals(center)) {
									centerControlValue += Piece.getPointValue(piece.getType());
								}
							}
						}
					}
				}
				return centerControlValue;
			}
	
			public static int getHighValuePieceThreatValue(DFSTreeNode node) {
				int highValueThreatValue = 0;
				for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					for (Move move : piece.getAllCaptureMoves(node.getGame())) {
						if (move instanceof MovementMove) {
							MovementMove movementMove = (MovementMove) move;
							Piece target = node.getGame().getBoard().getPieceAtPosition(movementMove.getTargetPosition());
							int pieceValue = Piece.getPointValue(target.getType());
							if (pieceValue >= Piece.getPointValue(PieceType.ROOK)) {
								highValueThreatValue += pieceValue;
							}
						}
					}
				}
				return highValueThreatValue;
			}
	
			public static int getBoardControlValue(DFSTreeNode node) {
				int boardControlValue = 0;
				for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					boardControlValue += piece.getAllMoves(node.getGame()).size();
				}
				return boardControlValue;
			}
	
			public static int getNumberOfPiecesMaxPlayerIsThreatening(DFSTreeNode node) {
				int numPiecesMaxPlayerIsThreatening = 0;
				for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					numPiecesMaxPlayerIsThreatening += piece.getAllCaptureMoves(node.getGame()).size();
				}
				return numPiecesMaxPlayerIsThreatening;
			}
		}
	
		public static class DefensiveHeuristics extends Object {
	
			public static int getKingSafetyValue(DFSTreeNode node) {
				int kingSafetyPenalty = 0;
				Piece kingPiece = node.getGame().getBoard().getPieces(getMaxPlayer(node), PieceType.KING).iterator().next();
				Coordinate kingPosition = node.getGame().getBoard().getPiecePosition(kingPiece);
	
				for (Direction direction : Direction.values()) {
					Coordinate neighborPosition = kingPosition.getNeighbor(direction);
					if (node.getGame().getBoard().isInbounds(neighborPosition) && node.getGame().getBoard().isPositionOccupied(neighborPosition)) {
						Piece nearbyPiece = node.getGame().getBoard().getPieceAtPosition(neighborPosition);
						if (nearbyPiece != null && kingPiece.isEnemyPiece(nearbyPiece)) {
							kingSafetyPenalty -= Piece.getPointValue(nearbyPiece.getType());
						}
					}
				}
				return Math.max(kingSafetyPenalty, -100);
			}
	
			public static int getPawnStructurePenalty(DFSTreeNode node) {
				int pawnPenalty = 0;
				for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					if (piece.getType() == PieceType.PAWN) {
						Coordinate pawnPosition = node.getGame().getBoard().getPiecePosition(piece);
						if (!isPieceProtected(node, pawnPosition)) {
							pawnPenalty -= 50;  // Penalize unprotected pawns
						}
					}
				}
				return pawnPenalty;
			}
	
			public static int getHighValuePieceProtectionValue(DFSTreeNode node) {
				int protectionValue = 0;
				for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					int pieceValue = Piece.getPointValue(piece.getType());
					if (pieceValue >= Piece.getPointValue(PieceType.ROOK)) {
						Coordinate piecePosition = node.getGame().getBoard().getPiecePosition(piece);
						if (!isPieceProtected(node, piecePosition)) {
							protectionValue -= pieceValue;
						}
					}
				}
				return protectionValue;
			}
	
			private static boolean isPieceProtected(DFSTreeNode node, Coordinate piecePosition) {
				for (Piece ally : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
					if (!ally.getCurrentPosition(node.getGame().getBoard()).equals(piecePosition) && ally.getAllMoves(node.getGame()).contains(piecePosition)) {
						return true;
					}
				}
				return false;
			}
	
			public static int getNumberOfMaxPlayersAlivePieces(DFSTreeNode node) {
				int numMaxPlayersPiecesAlive = 0;
				for (PieceType pieceType : PieceType.values()) {
					numMaxPlayersPiecesAlive += node.getGame().getBoard().getNumberOfAlivePieces(getMaxPlayer(node), pieceType);
				}
				return numMaxPlayersPiecesAlive;
			}
	
			public static int getClampedPieceValueTotalSurroundingMaxPlayersKing(DFSTreeNode node) {
				int maxPlayerKingSurroundingPiecesValueTotal = 0;
				Piece kingPiece = node.getGame().getBoard().getPieces(getMaxPlayer(node), PieceType.KING).iterator().next();
				Coordinate kingPosition = node.getGame().getBoard().getPiecePosition(kingPiece);
	
				for (Direction direction : Direction.values()) {
					Coordinate neighborPosition = kingPosition.getNeighbor(direction);
					if (node.getGame().getBoard().isInbounds(neighborPosition) && node.getGame().getBoard().isPositionOccupied(neighborPosition)) {
						Piece piece = node.getGame().getBoard().getPieceAtPosition(neighborPosition);
						int pieceValue = Piece.getPointValue(piece.getType());
						if (piece != null && kingPiece.isEnemyPiece(piece)) {
							maxPlayerKingSurroundingPiecesValueTotal -= pieceValue;
						} else if (piece != null && !kingPiece.isEnemyPiece(piece)) {
							maxPlayerKingSurroundingPiecesValueTotal += pieceValue;
						}
					}
				}
				return Math.max(maxPlayerKingSurroundingPiecesValueTotal, 0);
			}
	
			public static int getNumberOfPiecesThreateningMaxPlayer(DFSTreeNode node) {
				int numPiecesThreateningMaxPlayer = 0;
				for (Piece piece : node.getGame().getBoard().getPieces(getMinPlayer(node))) {
					numPiecesThreateningMaxPlayer += piece.getAllCaptureMoves(node.getGame()).size();
				}
				return numPiecesThreateningMaxPlayer;
			}
		}
	
		public static double getNonlinearPieceCombinationMaxPlayerHeuristicValue(DFSTreeNode node) {
			double multiPieceValueTotal = 0.0;
			double exponent = 1.5;
	
			for (PieceType pieceType : new PieceType[]{PieceType.BISHOP, PieceType.KNIGHT, PieceType.ROOK, PieceType.QUEEN}) {
				int numberOfPieces = node.getGame().getBoard().getNumberOfAlivePieces(getMaxPlayer(node), pieceType);
				multiPieceValueTotal += Math.pow(numberOfPieces, exponent);  // Nonlinear bonus for keeping pairs of pieces
			}
	
			double mobilityValue = 0.0;
			for (Piece piece : node.getGame().getBoard().getPieces(getMaxPlayer(node))) {
				mobilityValue += piece.getAllMoves(node.getGame()).size();
			}
	
			return multiPieceValueTotal + mobilityValue;  // Combine synergy bonus with mobility
		}
	}